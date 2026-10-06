package archtest

import (
	"go/parser"
	"go/token"
	"io/fs"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

const (
	module = "github.com/minimundosbr/animeGo-cli"
	root   = "../.." // module root, relative to this package
)

// rule describes what the packages under layer are allowed to import.
type rule struct {
	layer      string
	internal   []string // internal layers it may import - The layers that aren't setted here, will be blocked by default.
	thirdParty bool     // may import packages outside the module and the standard library
	stdDenied  []string // standard library prefixes it may not import
}

var rules = []rule{
	{layer: "internal/domain", stdDenied: []string{"net", "os/exec", "database"}},
	{layer: "internal/usecase", internal: []string{"internal/domain"}, stdDenied: []string{"net"}},
}

func TestDependencyRule(t *testing.T) {
	err := filepath.WalkDir(root, func(file string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(file, ".go") || strings.HasSuffix(file, "_test.go") {
			return nil
		}

		rel, err := filepath.Rel(root, filepath.Dir(file))
		if err != nil {
			return err
		}
		pkg := path.Join(module, filepath.ToSlash(rel))
		r, ok := ruleFor(pkg)
		if !ok {
			return nil
		}

		f, err := parser.ParseFile(token.NewFileSet(), file, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, spec := range f.Imports {
			imp, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				return err
			}
			if reason := r.violation(pkg, imp); reason != "" {
				t.Errorf("%s imports %s: %s", file, imp, reason)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func (r rule) violation(from, imp string) string {
	switch {
	case hasPathPrefix(imp, module):
		if hasPathPrefix(imp, from) {
			return ""
		}
		for _, allowed := range r.internal {
			if hasPathPrefix(imp, module+"/"+allowed) {
				return ""
			}
		}
		return r.layer + " may not import this layer"
	case isStandard(imp):
		for _, denied := range r.stdDenied {
			if hasPathPrefix(imp, denied) {
				return "standard library package not allowed in " + r.layer
			}
		}
	default:
		if !r.thirdParty {
			return r.layer + " may not import third-party packages"
		}
	}
	return ""
}

func ruleFor(pkg string) (rule, bool) {
	for _, r := range rules {
		if hasPathPrefix(pkg, module+"/"+r.layer) {
			return r, true
		}
	}
	return rule{}, false
}

// isStandard reports whether imp belongs to the standard library:
// its first path element has no dot (third-party paths start with a host name).
func isStandard(imp string) bool {
	first, _, _ := strings.Cut(imp, "/")
	return !strings.Contains(first, ".")
}

// hasPathPrefix reports whether p is prefix itself or a sub-package of it.
func hasPathPrefix(p, prefix string) bool {
	return p == prefix || strings.HasPrefix(p, prefix+"/")
}

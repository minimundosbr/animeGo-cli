type Anime struct {
	ID                int      `json:"anime_id"`
	MainTitle         string   `json:"main_title"`
	NativeTitle       string   `json:"native_title"`
	AlternativeTitles []string `json:"alternative_titles"`
	Synopsis          string   `json:"synopsis"`
	Seasons           []Season `json:"seasons"`
}
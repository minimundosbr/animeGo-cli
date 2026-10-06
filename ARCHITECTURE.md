## 1. `usecase` não importa nada de `net`

- **Decisão:** a regra de dependência bloqueia `net` inteiro em `usecase`, não só `net/http`.
- **Alternativa:** bloquear apenas `net/http`, que é o mínimo pedido pelo desafio.
- **Motivo:** toda rede fica atrás de uma porta. Bloquear `net` inteiro impede que tipos
  como `url.URL` ou `net.Error` apareçam nas portas ou nos casos de uso.
- **Consequência:** os adapters traduzem falhas de rede para erros do domínio
  (`ErrSourceUnavailable`), porque o `usecase` não consegue inspecionar `net.Error`.
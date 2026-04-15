# Clube do Livro

Aplicação web fechada para membros de um clube de leitura gerenciarem sugestões
de livros, votações, avaliações e encontros. Tema visual escuro (azul-noite)
por padrão.

Stack:

- **Go 1.23** (`net/http` + `chi`) servindo API e HTML renderizado no servidor.
- **HTMX** no front-end para interatividade sem JS customizado.
- **PostgreSQL** como banco de dados.
- **Docker Compose** para subir tudo com um único comando.

> **Nota sobre Templ:** a estrutura segue o layout pedido (`/internal/templ`).
> Os templates são escritos em `html/template` da stdlib, mantendo a mesma
> filosofia de Templ (SSR tipado) sem exigir etapa de `templ generate`.
> Migrar para `a-h/templ` é direto: basta substituir os arquivos `.gohtml` por
> `.templ` e regenerar.

## Rodando localmente (sem Docker)

Pré-requisitos: Go 1.23+, Postgres 14+.

```bash
cp .env.example .env
# ajuste DATABASE_URL etc em .env

make tidy      # baixa dependências
make run       # sobe o servidor em http://localhost:8080
```

As migrações rodam automaticamente na inicialização. Um usuário
administrador inicial é criado com as credenciais em `ADMIN_EMAIL` /
`ADMIN_PASSWORD` (padrão: `admin@clube.local` / `admin123`).

## Rodando com Docker

```bash
cp .env.example .env
make docker-up
```

A aplicação fica disponível em `http://localhost:8080` e o Postgres em
`localhost:5432`.

## Testes

```bash
make test
```

A suíte cobre serviços (com mocks de store), busca em Google Books / Open
Library via `httptest`, médias de avaliação, lógica de votação (abertura,
voto único, encerramento, vencedor determinístico) e renderização das
principais views (regressões de templates).

## Testando envio de e-mail localmente

Por padrão, com `SMTP_HOST` vazio, o `SMTPMailer` apenas registra o e-mail
no log (`[email mock] para=… assunto=… corpo=…`). Para inspecionar mensagens
em uma UI, suba um Mailpit local:

```yaml
# em docker-compose.yml
mailpit:
  image: axllent/mailpit
  ports: ["1025:1025", "8025:8025"]
```

E configure o serviço `app`:

```yaml
SMTP_HOST: mailpit
SMTP_PORT: "1025"
SMTP_FROM: clube@local
```

Acesse `http://localhost:8025` para visualizar os e-mails capturados.

## Estrutura

```
cmd/server           entrypoint principal
internal/handler     handlers HTTP por domínio
internal/service     lógica de negócio
internal/store       queries ao banco (interface + Postgres)
internal/model       structs de domínio
internal/templ       templates HTML (html/template)
internal/middleware  auth, logging, recovery
migrations           SQL numerado embutido via //go:embed
```

## Variáveis de ambiente

Consulte `.env.example`. As principais:

| Variável | Descrição |
|---|---|
| `PORT` | Porta HTTP (padrão 8080) |
| `DATABASE_URL` | DSN Postgres |
| `SMTP_HOST` / `SMTP_PORT` / `SMTP_USER` / `SMTP_PASS` / `SMTP_FROM` | SMTP para lembretes (vazio = mock no log) |
| `GOOGLE_BOOKS_API_KEY` | Opcional, aumenta a cota da API |
| `ADMIN_EMAIL` / `ADMIN_PASSWORD` / `ADMIN_NAME` | Seed do admin inicial |

## Funcionalidades

- Cadastro manual de membros (admin) — não há auto-registro.
- Login com `bcrypt` e sessão por cookie httpOnly opaco.
- Sugestão de livros com busca em Google Books (fallback: Open Library).
- Tags livres por livro.
- Rodadas de votação com voto único por membro e contagem oculta durante.
- Avaliações com todos os campos opcionais, indicador de spoiler e média.
- Encontros com pauta colaborativa, presença e lembrete SMTP 24h antes.

## Tema visual

Interface dark única, mesclando azul-marinho profundo e preto:

- Fundo: `#05070f` (quase preto com viés azul)
- Cards/seções: `#0e1730` (azul-marinho profundo)
- Sidebar: gradiente `#0f1d44 → #050912` (azul-escuro a quase-preto)
- Acentos: esmeralda, coral, dourado e azul `#5b8cff` para realces

Layout responsivo com sidebar colapsável (hamburger + backdrop) abaixo de
760px e ajustes adicionais abaixo de 480px.

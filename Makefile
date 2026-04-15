.PHONY: run test build migrate docker-up docker-down tidy

# Roda a aplicação localmente. Espera que o Postgres esteja acessível.
run:
	go run ./cmd/server

# Roda toda a suíte de testes.
test:
	go test ./... -count=1

# Compila o binário do servidor em ./bin/server
build:
	mkdir -p bin
	go build -o bin/server ./cmd/server

# Roda migrações manualmente contra o DATABASE_URL atual.
# A aplicação também roda as migrações automaticamente na inicialização.
migrate:
	go run ./cmd/server -migrate-only

# Sobe a stack inteira com Docker Compose.
docker-up:
	docker compose up --build

docker-down:
	docker compose down

tidy:
	go mod tidy

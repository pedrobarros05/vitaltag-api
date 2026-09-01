# VitalTag — API

Backend do **VitalTag**, sistema para compartilhamento de informações clínicas em situações de
emergência no ambiente de trabalho. O trabalhador cadastra suas informações clínicas e recebe um
QR Code com o link para uma página pública protegida por senha impressa junto ao código.

Projeto da **Disciplina Integradora de Desenvolvimento de Projetos** — Análise e Desenvolvimento
de Sistemas.

Frontend: **vitaltag-web** (repositório separado).

## Stack

- **Go 1.23** — API REST
- **chi** — roteamento e middlewares
- **pgx** — driver PostgreSQL
- **PostgreSQL 16** — banco de dados
- **golang-jwt** — tokens de sessão · **bcrypt** — hash de senhas
- **skip2/go-qrcode** — geração do QR Code
- **golang-migrate** — versionamento do schema

## Arquitetura

```
cmd/api           ponto de entrada
internal/handler      camada HTTP: rotas, parsing e validação de entrada
internal/service      regras de negócio (senhas distintas, bloqueio, exclusão em cascata)
internal/repository   acesso ao PostgreSQL
internal/model        entidades do domínio
internal/middleware   autenticação JWT, rate limit, logs
migrations            scripts de schema versionados
```

## Como rodar (Fase 2)

```bash
cp .env.example .env
docker compose up -d db
migrate -path migrations -database "$DATABASE_URL" up
go run ./cmd/api
```

## Endpoints previstos

| Método | Rota | Descrição |
|---|---|---|
| POST | `/api/v1/auth/registro` | cria conta (RF01) |
| POST | `/api/v1/auth/login` | autentica e emite token (RF02) |
| GET | `/api/v1/perfil` | retorna o perfil do usuário autenticado (RF11) |
| POST | `/api/v1/perfil` | cria/atualiza informações clínicas (RF04, RF05, RF12) |
| GET | `/api/v1/perfil/qrcode` | devolve o QR Code em PNG/SVG (RF06) |
| DELETE | `/api/v1/perfil` | exclui o registro completo (RF10) |
| POST | `/api/v1/publico/{slug}` | valida a senha pública e retorna os dados (RF08, RF09) |

## Documentação

A modelagem completa (requisitos, User Stories, EAP, roadmap, protótipo e modelo de dados) está
na entrega da Fase 1 da disciplina.

# anti-yt

[日本語](README.md) | English

## Architecture

```mermaid
flowchart LR
    user["User"]
    cf["Cloudflare Tunnel"]
    nginx["Nginx"]
    front["Preact"]
    backend["Backend"]
    pg["PostgreSQL"]
    redis["Redis"]
    yt["YouTube Data API"]
    user -->|HTTPS| cf
    cf --> nginx
    nginx -->|Static serving| front
    nginx -->|/api/v1/*| backend
    backend --> pg
    backend --> redis
    backend -->|Data API| yt
````

## Setup

We recommend opening the project with VSCode Devcontainer.

Copy `.devcontainer/.env.example` to create `.devcontainer/.env`, then set each value.

| Variable | Description |
| --- | --- |
| `PORT` | Backend port |
| `ENV` | Runtime environment (development or production) |
| `OIDC_GOOGLE_CLIENT_ID` |  |
| `OIDC_GOOGLE_CLIENT_SECRET` |  |
| `DATABASE_URL` | PostgreSQL |
| `SERVER_URL` | Backend URL |
| `FRONTEND_URL` | Frontend URL |
| `YOUTUBE_DATA_API_KEY` | YouTube Data API |
| `ADMIN_API_KEY` | API key for anti-yt administrators. You can set it to whatever you like. |
| `GEMINI_API_KEY` | Gemini API |
| `DISCORD_WEBHOOK_URL` | Discord Webhook |
| `REDIS_URL` | Redis |
| `OTEL_EXPORTER_OTLP_PROTOCOL` |  |
| `OTEL_EXPORTER_OTLP_ENDPOINT` |  |
| `OTEL_EXPORTER_OTLP_HEADERS` |  |

Generate the key files for JWT in `.devcontainer/`.

```sh
# Private key
openssl genpkey -algorithm Ed25519 -out .devcontainer/jwt_private.pem

# Public key
openssl pkey -in .devcontainer/jwt_private.pem -pubout -out .devcontainer/jwt_public.pem
```

Generate a self-signed certificate for Nginx in `.devcontainer/`.

```sh
openssl req -x509 -newkey ec -pkeyopt ec_paramgen_curve:prime256v1 -nodes \
  -keyout .devcontainer/server.key -out .devcontainer/server.crt \
  -days 365 -subj '/CN=localhost'
```

After opening the devcontainer, run the following commands to install dependencies and start the development servers.

Backend:

```sh
cd backend
air
```

Frontend:

```sh
cd frontend
npm i
npm run dev
```

## API Definition

The OpenAPI definition is at [shared/api/v1/openapi.yaml](shared/api/v1/openapi.yaml).
The backend uses oapi-codegen and the frontend uses orval to generate code from OpenAPI.
After starting the backend server, you can view the API definition graphically at `/api/v1/swagger` (only when the `ENV` environment variable is `development`).

## DB Definition

The DB definition is at [backend/docs/schema](backend/docs/schema/).
It is generated using tbls.

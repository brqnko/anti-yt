# anti-yt

## 環境構築

VSCodeのDevcontainerを使って開くことをおすすめします。

.devcontainer/.env.exampleをコピーして.devcontainer/.envを作成し、各値を設定してください。

| 変数名 | 説明 |
| --- | --- |
| `PORT` | バックエンドのポート |
| `ENV` | 実行環境（developmentかproduction) |
| `OIDC_GOOGLE_CLIENT_ID` |  |
| `OIDC_GOOGLE_CLIENT_SECRET` |  |
| `DATABASE_URL` | PostgreSQL |
| `SERVER_URL` | バックエンドのURL |
| `FRONTEND_URL` | フロントエンドのURL |
| `YOUTUBE_DATA_API_KEY` | YouTube Data API |
| `ADMIN_API_KEY` | anti-yt管理者用APIキー。自分の好きなものにできます。 |
| `GEMINI_API_KEY` | Gemini API |
| `DISCORD_WEBHOOK_URL` | Discord Webhook |
| `REDIS_URL` | Redis |
| `OTEL_EXPORTER_OTLP_PROTOCOL` |  |
| `OTEL_EXPORTER_OTLP_ENDPOINT` |  |
| `OTEL_EXPORTER_OTLP_HEADERS` |  |

JWT用の鍵ファイルを.devcontainer/に生成してください。

```sh
# 秘密鍵
openssl genpkey -algorithm Ed25519 -out .devcontainer/jwt_private.pem

# 公開鍵
openssl pkey -in .devcontainer/jwt_private.pem -pubout -out .devcontainer/jwt_public.pem
```

PostgreSQL用のパスワードファイルを.devcontainer/password.txtに作成してください。

Nginx用の自己署名証明書を.devcontainer/に生成してください。

```sh
openssl req -x509 -newkey ec -pkeyopt ec_paramgen_curve:prime256v1 -nodes \
  -keyout .devcontainer/server.key -out .devcontainer/server.crt \
  -days 365 -subj '/CN=localhost'
```

devcontainerを開いた後、以下のコマンドで依存関係のインストールと開発サーバーの起動を行います。

バックエンド:

```sh
cd backend
air
```

フロントエンド:

```sh
cd frontend
npm i
npm run dev
```

## API定義

[shared/api/v1/openapi.yaml](shared/api/v1/openapi.yaml)にOpenAPI定義があります。
Backendはoapi-codegen、Frontendはorvalを使ってOpenAPIからwいsdーを行いますい。
Backendサーバー起動後、/api/v1/swaggerからAPI定義をグラフィカルに見れます(ENVの環境変数がdevelopmentの場合のみ)。

## DB定義

[backend/docs/schema](backend/docs/schema/)にDB定義があります。
tblsを用いて生成しています。

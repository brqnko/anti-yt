# anti-yt

anti-ytのソースコードを管理するリポジトリです。

以下のドキュメントがあります:

- [backendの設計](backend/docs)
- [frontendの設計](frontend/docs)
- [OpenAPI定義](shared/api/v1/openapi.yaml)
- [開発ブログ](https://blog.brqnko.rs/youtube-addiction)

## 環境構築

VSCodeのDevcontainerを使って開くことをおすすめします。
devcontainerを開くとバックエンド・フロントエンドの開発環境が統合された`full`コンテナとPostgresコンテナが起動し、VSCodeは`full`にアタッチされます。
バックエンドとフロントエンドが1つのコンテナにまとまっているためコンテナを切り替える必要がなく、node_modulesやLSPのキャッシュもコンテナ内に閉じるためホストOSを汚しません。

.devcontainer/.env.exampleをコピーして.devcontainer/.envを作成し、各値を設定してください。

| 変数名 | 説明 |
| --- | --- |
| `PORT` | バックエンドのポート番号 |
| `ENV` | 実行環境（`development` / `production`） |
| `OIDC_GOOGLE_CLIENT_ID` | Google OIDCのクライアントID |
| `OIDC_GOOGLE_CLIENT_SECRET` | Google OIDCのクライアントシークレット |
| `DATABASE_URL` | PostgreSQLの接続URL |
| `SERVER_URL` | バックエンドのURL |
| `FRONTEND_URL` | フロントエンドのURL |
| `YOUTUBE_DATA_API_KEY` | YouTube Data APIのキー |
| `ADMIN_API_KEY` | 管理者用APIキー |
| `GEMINI_API_KEY` | Gemini APIのキー |
| `DISCORD_WEBHOOK_URL` | Discord WebhookのURL |
| `REDIS_URL` | RedisのURL |

JWT用の鍵ファイルを.devcontainer/に生成してください。

```sh
# 秘密鍵の生成
openssl genpkey -algorithm Ed25519 -out .devcontainer/jwt_private.pem

# 公開鍵の生成
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

バックエンド起動後、`/api/v1/swagger`からOpenAPI定義を閲覧できます。

フロントエンド:

```sh
cd frontend
npm i
npm run dev
```

## デプロイ

[`backend/Dockerfile`](backend/Dockerfile)と[`frontend/Dockerfile`](frontend/Dockerfile)でそれぞれのイメージをビルドできます。

[`compose.yml`](compose.yml)でデプロイできますが、PostgreSQLとNginxは外部で管理する構成になっています。
1つのサーバー上でPostgreSQLとNginxを共有し、複数のサービスがそれらを利用する形です。
フロントエンドはSSG + PWAのため、ビルド済みの静的ファイルをNginxで配信します。
APIへのリクエストはサブパス方式で振り分けてます。これによりCORS設定が不要になります。

# frontend

これは anti-yt のフロントエンド設計に関するドキュメントです。

フロントエンドは Google Stitch + Claude Code で実装しました。

## 技術スタック

- Preact + preact-iso によるSPAルーティング / プリレンダリング
- SWR によるキャッシュ・再検証
- axios + orval による型安全なAPIクライアント生成
- Tailwind CSS によるスタイリング
- i18next / react-i18next による国際化 (ja / en)
- Vite + vite-plugin-pwa によるビルド・PWA化

## APIクライアント

- `shared/api/v1/openapi.yaml` を入力として、orval で `src/api/generated` 配下に axios クライアントを生成します
- 生成ファイルは直接編集しません。スキーマを変更した場合は `npm run generate` で再生成します
- 共通の挙動 (Cookie 送信・CSRFトークン・デバイスフィンガープリント・タイムゾーンヘッダ) は `src/api/axios-instance.ts` の interceptor に集約します
- 401 発生時は `/api/v1/auth/refresh` を呼び、単一の refresh を並列リクエストで共有するキューを設けます
- refresh 失敗時は `auth:logout` カスタムイベントを発火し、AuthContext で購読してログアウト処理に繋げます

## 状態管理

### SWR

- 読み取りクエリは SWR を通じて取得し、`revalidateOnFocus` / `revalidateOnReconnect` を有効化します
- 複数画面で共有されるキーは `src/api/cache-keys.ts` に集約し、タイプミスや重複を避けます
- 書き込み後の再検証は `mutate` を用い、関連するキャッシュキーに対して明示的に行います

## ルーティング

- preact-iso の `Router` / `Route` を使用し、ページは `lazy()` による動的importで分割します
- 認証が必要な画面は `ProtectedRoute` でラップし、未認証時はログインフローへ誘導します
- ページ単位のディレクトリには `index.tsx` と、そのページ固有のフック・コンポーネントを同居させます

## 認証フロー

- OIDC によるソーシャルログイン後、バックエンドから Cookie で JWT を受け取ります
- CSRF トークンは Cookie から読み取り、`x-csrf-token` ヘッダに載せて送信します
- デバイスフィンガープリントは FingerprintJS で生成し、`X-Device-Fingerprint` ヘッダで送信します
- refresh / logout は axios インスタンスの interceptor と AuthContext が協調して処理します

## ビルド・配信

- `vite build` により静的ファイルを生成し、`prerender` 関数を通じて主要ルートを事前レンダリングします
- PWA 化は `vite-plugin-pwa` を使用し、Service Worker によるキャッシュを有効にします
- Docker イメージは `Dockerfile` で静的配信サーバとして構成します

# Frontend API・認証アーキテクチャ

## 概要

本フロントエンド(Preact)では、認証トークンを **HTTP-only Cookie** で管理し、Axiosインターセプターによる自動トークンリフレッシュを実装している。

## ファイル構成

| ファイル | 役割 |
|---|---|
| `src/api/axios-instance.ts` | Axiosインスタンス生成、リクエスト/レスポンスインターセプター |
| `src/api/mutator.ts` | Orval用カスタムミューテーター（AbortController対応） |
| `src/api/generated/auth.ts` | 自動生成された認証系APIクライアント |
| `src/api/generated/user.ts` | 自動生成されたユーザー系APIクライアント |
| `src/contexts/AuthContext.tsx` | 認証状態管理（Preact Context） |
| `src/components/ProtectedRoute.tsx` | 認証必須ページのガード |
| `src/utils/cookie.ts` | Cookie読み取りユーティリティ |

## トークンの保存場所

| トークン | 保存先 | JSからアクセス可能か |
|---|---|---|
| Access Token | HTTP-only Cookie（サーバー設定） | 不可 |
| Refresh Token | HTTP-only Cookie（サーバー設定） | 不可 |
| CSRF Token | 通常のCookie（`csrf_token`） | 可能（`getCookie("csrf_token")`） |

- `localStorage` / `sessionStorage` にはトークンを **一切保存しない**
- `withCredentials: true` により、全リクエストにCookieが自動付与される

## リクエストインターセプター

`axios-instance.ts` L24-36で全リクエストに以下のヘッダーを付与:

1. **`X-Device-Fingerprint`** — FingerprintJS v5でブラウザ固有IDを生成し付与（初回ロード後キャッシュ）
2. **`x-csrf-token`** — Cookieから`csrf_token`を読んでヘッダーに設定（CSRF対策）

## レスポンスインターセプター（トークンリフレッシュ）

`axios-instance.ts` L38-91で401レスポンスを検知し、自動的にトークンリフレッシュを行う。

### フロー

1. APIが `401` を返す
2. 以下の条件に該当する場合はリフレッシュせずそのままreject:
   - 既にリトライ済み（`originalRequest._retry === true`）
   - リフレッシュエンドポイント自体のリクエスト（無限ループ防止）
   - 401以外のエラー
3. `POST /api/v1/auth/refresh` を呼び出し
4. 成功 → 元のリクエストをリトライ
5. 失敗 → `auth:logout` カスタムイベントを発火し、AuthContextがログアウト処理

### キューイング（同時リフレッシュ防止）

```
isRefreshing = true の間に来た401リクエストは failedQueue に格納
→ リフレッシュ完了後、キュー内の全リクエストをresolve/rejectして再実行
```

これにより複数の同時401レスポンスがあっても、リフレッシュAPIは1回だけ呼ばれる。

## AuthContext（認証状態管理）

`src/contexts/AuthContext.tsx` で以下の状態を管理:

| State | 説明 |
|---|---|
| `isLoading` | 認証チェック中 |
| `isAuthenticated` | 認証済みか |
| `error` | 認証チェックエラー |
| `sessionExpired` | セッション期限切れフラグ |

### 主要メソッド

- **`checkAuth()`** — `GET /api/v1/users/me` を呼び、200なら認証済み、401なら未認証（エラーにしない）
- **`logout()`** — `POST /api/v1/auth/logout` → 状態クリア → `/` にリダイレクト
- **`refreshAuth()`** — `checkAuth()` のエイリアス

### イベントリスナー

インターセプターが発火する `auth:logout` カスタムイベントを監視:
- `reason: "session_expired"` → `sessionExpired = true` に設定
- ProtectedRouteが `/?expired=1` にリダイレクト

## 認証フロー全体像

### 初回アクセス

1. `AuthProvider` マウント → `checkAuth()` 実行
2. `GET /api/v1/users/me` → 200ならログイン済み / 401なら未ログイン

### Googleログイン

1. `/api/v1/auth/google` にリダイレクト
2. バックエンドがGoogle OAuthフロー処理
3. 成功時、バックエンドがHTTP-only CookieにAccess/Refresh Tokenを設定
4. フロントに戻り、`checkAuth()` で認証状態を確認

### トークン期限切れ

1. APIが401を返す
2. レスポンスインターセプターが `POST /api/v1/auth/refresh` を自動実行
3. 成功 → 元のリクエストをリトライ（ユーザーは気づかない）
4. 失敗 → セッション期限切れとしてログアウト → `/?expired=1` にリダイレクト

### ログアウト

1. `POST /api/v1/auth/logout` → サーバー側でセッション無効化
2. フロント側の認証状態をクリア
3. `/` にリダイレクト

## セキュリティ対策

- **XSS対策**: HTTP-only CookieによりJSからトークンにアクセス不可
- **CSRF対策**: CSRFトークンをヘッダーに付与
- **デバイスフィンガープリント**: 不正アクセス検知用
- **リフレッシュループ防止**: リフレッシュURL自体の401はスキップ
- **キューイング**: 同時リフレッシュによるrace condition防止

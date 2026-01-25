# ER図

## 論理設計

```mermaid
erDiagram
    m_user {
        bigint m_user_id PK "連番"
        bigint m_user_authorization_id FK
        varchar(32) display_name "word charactor. 数字では始まらない"
        varchar(2) language_code "ISO 639-1"
        int daily_screen_time_seconds "一日に見れる制限時間. 指定しない場合は86401秒"
        timestamptz joined_at "ビジネス上のアカウント作成時間"
        timestamptz created_at
        timestamptz updated_at
        uuid public_id "Web API用. Base64 URL safeでエンコードする"
    }

    m_user_authorization {
        bigint m_user_authorization_id PK
        varchar(256) issuer "https://accounts.google.com"
        varchar(256) sub
        timestamptz last_logged_in_at
        timestamptz created_at
        timestamptz updated_at
        uuid public_id
    }

    m_refresh_token {
        bigint m_refresh_token_id PK
        bigint m_user_authorization_id FK
        varchar(64) token_hash "sha256(base64url(safeRandom(32bit)))"
        int generation
        varchar(64) ip_address "RFC 5952"
        varchar(32) device_fingerprint "FingerprintJSのvisitorId"
        varchar(512) user_agent "goのuser_agentパッケージを使用"
        varchar(2) country_code "Cloudflare Tunnelから取得"
        varchar(128) city_name
        varchar(64) browser_name
        varchar(32) device_type
        timestamptz expires_at
        timestamptz created_at
        timestamptz updated_at
        uuid public_id
    }

    h_user {
        bigint h_user_id PK
        bigint m_user_authorization_id FK
        varchar(32) display_name "m_userと同じ"
        varchar(2) language_code
        int daily_screen_time_seconds
        timestamptz joined_at
        timestamptz left_at
        int leave_reason_code
        timestamptz created_at
        timestamptz updated_at
        uuid public_id
    }

    m_user_screen_time_range {
        bigint m_user_screen_time_range_id PK
        bigint m_user_id FK
        time screen_time_range_start
        time screen_time_range_end
        timestamptz created_at
        timestamptz updated_at
        uuid public_id
    }

    m_user_subscribing_channel {
        bigint m_user_subscribing_channel_id PK
        bigint m_user_id FK
        bigint m_channel_id FK
        timestamptz subscribed_at
        timestamptz created_at
        timestamptz updated_at
        uuid public_id
    }

    m_channel {
        bigint m_channel_id PK
        varchar(32) external_id
        varchar(64) external_display_name
        varchar(64) external_custom_id
        varchar(512) external_icon_url
        varchar(1024) external_description
        bigint external_subscribers_count
        timestamptz external_created_at
        timestamptz fetched_at
        timestamptz created_at
        timestamptz updated_at
        uuid public_id
    }

    m_video {
        bigint m_video_id PK
        bigint m_channel_id FK
        varchar(16) external_id
        varchar(128) external_title
        varchar(8192) external_description
        bigint external_like_count
        bigint external_watch_count
        timestamptz external_created_at
        int external_length_seconds
        timestamptz fetched_at
        timestamptz created_at
        timestamptz updated_at
        uuid public_id
    }

    m_comment {
        bigint m_comment_id PK
        bigint m_video_id FK
        varchar(512) external_id
        varchar(8192) external_content
        bigint external_like_count
        varchar(32) external_user_id
        varchar(64) external_user_display_name
        varchar(64) external_user_custom_id
        bool external_edited_flg
        timestamptz external_created_at
        timestamptz fetched_at
        timestamptz created_at
        timestamptz updated_at
        uuid public_id
    }

    t_video_watch {
        bigint t_video_watch_id PK
        bigint m_user_id FK
        bigint m_video_id FK
        timestamptz watch_start_at "(m_user_id, m_video_id)でEXCLUDE制約あり"
        timestamptz watch_end_at
        int watch_position_seconds
        timestamptz created_at
        timestamptz updated_at
        uuid public_id
    }

    h_search {
        bigint h_search_id PK
        bigint m_user_id FK
        varchar(256) search_keyword
        timestamptz searched_at
        timestamptz created_at
        timestamptz updated_at
    }

    s_monthly_video_watch {
        bigint s_monthly_video_watch_id PK
        bigint m_user_id FK
        varchar(128) ai_summary_title
        varchar(4096) ai_summary_description
        varchar(128) ai_model
        timestamptz generated_at
        date target_month
        bigint w_monthly_video_watch_id FK
        timestamptz created_at
        timestamptz updated_at
        uuid public_id
    }

    w_monthly_video_watch {
        bigint w_monthly_video_watch_id PK
        bigint m_user_id FK
        int batch_status_code "デフォルトは1"
        varchar(128) ai_model
        timestamptz started_at
        timestamptz finished_at
        date target_month
        varchar(128) fail_reason
        timestamptz created_at
        timestamptz updated_at
    }

    m_playlist {
        bigint m_playlist_id PK
        bigint m_user_id FK
        int visibility_code "とりあえず非公開のみで実装"
        varchar(128) playlist_title
        int playlist_code "とりあえず9で運用。将来的にカテゴリー分けなどが必要になるかもしれない。"
        int video_count
        int playlist_total_video_length_seconds
        timestamptz created_at
        timestamptz updated_at
        uuid public_id
    }

    m_playlist_video {
        bigint m_playlist_video_id PK
        bigint m_playlist_id FK
        bigint m_video_id FK
        bigint playlist_position "ギャップ"
        timestamptz created_at
        timestamptz updated_at
    }
    
    t_jti_blacklist {
        uuid jti
        timestamptz expires_at
    }

    m_channel ||--o{ m_user_subscribing_channel: ""
    m_user ||--o{ m_user_subscribing_channel: ""
    m_video ||--o{ t_video_watch: ""
    m_user ||--o{ t_video_watch: ""
    m_video ||--o{ m_comment: ""
    m_user ||--o{ m_user_screen_time_range: ""
    m_channel ||--o{ m_video: ""
    m_user ||--o{ h_search: ""
    m_user ||--o{ s_monthly_video_watch: ""
    s_monthly_video_watch ||--|| w_monthly_video_watch: "成功したら紐付ける"
    m_user_authorization ||--|| m_user: "現在は1Googleアカウント=1ユーザー"
    m_user_authorization ||--|| h_user: "現在は1Googleアカウント=1ユーザー"
    m_refresh_token }o--|| m_user_authorization: ""
    s_monthly_video_watch }o--|| m_user: ""
    m_playlist }o--|| m_user: ""
    m_video ||--o{ m_playlist_video: ""
    m_playlist ||--o{ m_playlist_video: ""
```

## インデックス

| テーブル名                      | インデックス名                          | カラム                                                       |
|----------------------------|----------------------------------|-----------------------------------------------------------|
| m_user                     | uk_1_m_user                      | public_id                                                 |
| m_user_authorization       | uk_1_m_user_authorization        | issuer, sub                                               |
| m_user_authorization       | uk_2_m_user_authorization        | public_id                                                 |
| m_refresh_token            | idx_1_m_refresh_token            | expires_at                                                |
| m_refresh_token            | uk_1_m_refresh_token             | token_hash                                                |
| m_refresh_token            | uk_2_m_refresh_token             | public_id                                                 |
| h_user                     | uk_1_h_user                      | public_id                                                 |
| m_user_screen_time_range   | idx_1_m_user_screen_time_range   | m_user_id                                                 |
| m_user_screen_time_range   | uk_1_m_user_screen_time_range    | public_id                                                 |
| m_user_subscribing_channel | idx_1_m_user_subscribing_channel | m_user_id                                                 |
| m_user_subscribing_channel | uk_1_m_user_subscribing_channel  | m_user_id, m_channel_id                                   |
| m_user_subscribing_channel | uk_2_m_user_subscribing_channel  | public_id                                                 |
| m_channel                  | uk_1_m_channel                   | public_id                                                 |
| m_channel                  | uk_2_m_channel                   | external_id                                               |
| m_channel                  | uk_3_m_channel                   | external_custom_id                                        |
| m_video                    | idx_1_m_video                    | m_channel_id                                              |
| m_video                    | uk_1_m_video                     | public_id                                                 |
| m_video                    | uk_2_m_video                     | external_id                                               |
| m_comment                  | idx_1_m_comment                  | m_video_id                                                |
| m_comment                  | uk_1_m_comment                   | public_id                                                 |
| m_comment                  | uk_2_m_comment                   | external_id                                               |
| t_video_watch              | uk_1_t_video_watch               | public_id                                                 |
| t_video_watch              | idx_1_t_video_watch              | m_user_id                                                 |
| h_search                   | idx_1_h_search                   | m_user_id                                                 |
| s_monthly_video_watch      | uk_1_s_monthly_video_watch       | public_id                                                 |
| s_monthly_video_watch      | uk_2_s_monthly_video_watch       | m_user_id, target_month                                   |
| m_playlist                 | uk_1_m_playlist                  | public_id                                                 |
| m_playlist                 | idx_1_m_playlist                 | m_user_id, playlist_visibility, playlist_code, created_at |
| m_playlist_video           | uk_1_m_playlist_video            | m_playlist_id, m_video_id                                 |
| t_jti_blacklist            | uk_1_t_jti_blacklist             | jti                                                       |
| t_jti_blacklist            | idx_1_t_jti_blacklist            | expires_at                                                |
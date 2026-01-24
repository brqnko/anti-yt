-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS btree_gist;

CREATE TABLE m_user
(
    m_user_id                 bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    m_user_authorization_id   bigint      NOT NULL,

    display_name              VARCHAR(32) NOT NULL,
    language_code             VARCHAR(2)  NOT NULL,

    daily_screen_time_seconds INTEGER     NOT NULL DEFAULT 86401,

    joined_at                 timestamptz NOT NULL DEFAULT current_timestamp,

    created_at                timestamptz NOT NULL DEFAULT current_timestamp,
    updated_at                timestamptz NOT NULL DEFAULT current_timestamp,
    public_id                 uuid        NOT NULL DEFAULT uuidv7()
);

CREATE UNIQUE INDEX uk_1_m_user ON m_user (public_id);

CREATE TABLE m_user_authorization
(
    m_user_authorization_id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

    issuer                  VARCHAR(256) NOT NULL,
    sub                     VARCHAR(256) NOT NULL,
    email_address           VARCHAR(256) NOT NULL,

    last_logged_in_at       timestamptz  NOT NULL DEFAULT current_timestamp,

    created_at              timestamptz  NOT NULL DEFAULT current_timestamp,
    updated_at              timestamptz  NOT NULL DEFAULT current_timestamp
);

CREATE UNIQUE INDEX uk_1_m_user_authorization ON m_user_authorization (issuer, sub);
CREATE UNIQUE INDEX uk_2_m_user_authorization ON m_user_authorization (email_address);

CREATE TABLE m_refresh_token
(
    m_refresh_token_id      bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    m_user_authorization_id bigint       NOT NULL,

    token_hash              VARCHAR(256) NOT NULL,
    generation              INTEGER      NOT NULL DEFAULT 1,
    ip_address              VARCHAR(64)  NOT NULL,
    device_fingerprint      VARCHAR(32)  NOT NULL,
    user_agent              VARCHAR(512) NOT NULL,
    country_code            VARCHAR(2)   NOT NULL,
    city_name               VARCHAR(128) NOT NULL,
    browser_name            VARCHAR(64)  NOT NULL,
    device_type             VARCHAR(32)  NOT NULL,

    expires_at              timestamptz  NOT NULL,

    created_at              timestamptz  NOT NULL DEFAULT current_timestamp,
    updated_at              timestamptz  NOT NULL DEFAULT current_timestamp,

    FOREIGN KEY (m_user_authorization_id)
        REFERENCES m_user_authorization (m_user_authorization_id)
        ON DELETE CASCADE
);

CREATE UNIQUE INDEX uk_1_m_refresh_token ON m_refresh_token (token_hash);
CREATE INDEX idx_1_m_refresh_token ON m_refresh_token (expires_at);

CREATE TABLE h_user
(
    h_user_id                 bigint PRIMARY KEY,
    m_user_authorization_id   bigint      NOT NULL,

    display_name              VARCHAR(32) NOT NULL,
    language_code             VARCHAR(2)  NOT NULL,

    daily_screen_time_seconds INTEGER     NOT NULL,

    joined_at                 timestamptz NOT NULL,

    left_at                   timestamptz NOT NULL DEFAULT current_timestamp,
    leave_reason_code         INTEGER     NOT NULL,

    created_at                timestamptz NOT NULL DEFAULT current_timestamp,
    updated_at                timestamptz NOT NULL DEFAULT current_timestamp,
    public_id                 uuid        NOT NULL DEFAULT uuidv7()
);

CREATE UNIQUE INDEX uk_1_h_user ON h_user (public_id);

CREATE TABLE m_user_screen_time_range
(
    m_user_screen_time_range_id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    m_user_id                   bigint      NOT NULL,

    screen_time_range_start     time        NOT NULL,
    screen_time_range_end       time        NOT NULL,

    created_at                  timestamptz NOT NULL DEFAULT current_timestamp,
    updated_at                  timestamptz NOT NULL DEFAULT current_timestamp,

    public_id                   uuid        NOT NULL DEFAULT uuidv7()
);

CREATE UNIQUE INDEX uk_1_m_user_screen_time_range ON m_user_screen_time_range (public_id);
CREATE INDEX idx_1_m_user_screen_time_range ON m_user_screen_time_range (m_user_id);

CREATE TABLE m_user_subscribing_channel
(
    m_user_subscribing_channel_id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    m_user_id                     bigint      NOT NULL,
    m_channel_id                  bigint      NOT NULL,

    subscribed_at                 timestamptz NOT NULL DEFAULT current_timestamp,

    created_at                    timestamptz NOT NULL DEFAULT current_timestamp,
    updated_at                    timestamptz NOT NULL DEFAULT current_timestamp,

    public_id                     uuid        NOT NULL DEFAULT uuidv7()
);

CREATE UNIQUE INDEX uk_1_m_user_subscribing_channel ON m_user_subscribing_channel (m_user_id, m_channel_id);
CREATE UNIQUE INDEX uk_2_m_user_subscribing_channel ON m_user_subscribing_channel (public_id);
CREATE INDEX idx_1_m_user_subscribing_channel ON m_user_subscribing_channel (m_user_id);

CREATE TABLE m_channel
(
    m_channel_id               bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

    external_id                VARCHAR(32)   NOT NULL,
    external_display_name      VARCHAR(64)   NOT NULL,
    external_custom_id         VARCHAR(64)   NOT NULL,
    external_icon_url          VARCHAR(512)  NOT NULL,
    external_description       VARCHAR(1024) NOT NULL,
    external_subscribers_count bigint        NOT NULL,
    external_created_at        timestamptz   NOT NULL,

    fetched_at                 timestamptz   NOT NULL DEFAULT current_timestamp,

    created_at                 timestamptz   NOT NULL DEFAULT current_timestamp,
    updated_at                 timestamptz   NOT NULL DEFAULT current_timestamp,

    public_id                  uuid          NOT NULL DEFAULT uuidv7()
);

CREATE UNIQUE INDEX uk_1_m_channel ON m_channel (public_id);
CREATE UNIQUE INDEX uk_2_m_channel ON m_channel (external_id);
CREATE UNIQUE INDEX uk_3_m_channel ON m_channel (external_custom_id);

CREATE TABLE m_video
(
    m_video_id           bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

    m_channel_id         bigint        NOT NULL,

    external_id          VARCHAR(16)   NOT NULL,
    external_title       VARCHAR(128)  NOT NULL,
    external_description VARCHAR(8192) NOT NULL,
    external_like_count  bigint        NOT NULL,
    external_watch_count bigint        NOT NULL,

    length_seconds       bigint        NOT NULL,

    fetched_at           timestamptz   NOT NULL DEFAULT current_timestamp,

    created_at           timestamptz   NOT NULL DEFAULT current_timestamp,
    updated_at           timestamptz   NOT NULL DEFAULT current_timestamp,

    public_id            uuid          NOT NULL DEFAULT uuidv7()
);

CREATE INDEX idx_1_m_video ON m_video (m_channel_id);
CREATE UNIQUE INDEX uk_1_m_video ON m_video (public_id);
CREATE UNIQUE INDEX uk_2_m_video ON m_video (external_id);

CREATE TABLE m_comment
(
    m_comment_id               bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

    m_video_id                 bigint        NOT NULL,

    external_id                VARCHAR(512)  NOT NULL,
    external_content           VARCHAR(8192) NOT NULL,
    external_like_count        bigint        NOT NULL,
    external_user_id           VARCHAR(32)   NOT NULL,
    external_user_display_name VARCHAR(64)   NOT NULL,
    external_user_custom_id    VARCHAR(64)   NOT NULL,
    external_edited_flg        BOOLEAN       NOT NULL,
    external_created_at        timestamptz   NOT NULL,

    fetched_at                 timestamptz   NOT NULL DEFAULT current_timestamp,

    created_at                 timestamptz   NOT NULL DEFAULT current_timestamp,
    updated_at                 timestamptz   NOT NULL DEFAULT current_timestamp,

    public_id                  uuid          NOT NULL DEFAULT uuidv7()
);

CREATE INDEX idx_1_m_comment ON m_comment (m_video_id);
CREATE UNIQUE INDEX uk_1_m_comment ON m_comment (public_id);
CREATE UNIQUE INDEX uk_2_m_comment ON m_comment (external_id);

CREATE TABLE t_video_watch
(
    t_video_watch_id       bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

    m_user_id              bigint      NOT NULL,
    m_video_id             bigint      NOT NULL,

    watch_start_at         timestamptz NOT NULL DEFAULT current_timestamp,
    watch_end_at           timestamptz NOT NULL DEFAULT '9999-12-31 23:59:59UTC',
    watch_position_seconds int         NOT NULL DEFAULT 0,

    created_at             timestamptz NOT NULL DEFAULT current_timestamp,
    updated_at             timestamptz NOT NULL DEFAULT current_timestamp,

    CONSTRAINT excl_1_t_video_watch EXCLUDE USING gist (
        m_user_id WITH =,
        m_video_id WITH =,
        tstzrange(watch_start_at, watch_end_at) WITH &&
        ),

    public_id              uuid        NOT NULL DEFAULT uuidv7()
);

CREATE UNIQUE INDEX uk_1_t_video_watch ON t_video_watch (public_id);
CREATE INDEX idx_1_t_video_watch ON t_video_watch (m_user_id);

CREATE TABLE h_search
(
    h_search_id    bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

    m_user_id      bigint       NOT NULL,
    search_keyword VARCHAR(256) NOT NULL,
    searched_at    timestamptz  NOT NULL DEFAULT current_timestamp,

    created_at     timestamptz  NOT NULL DEFAULT current_timestamp,
    updated_at     timestamptz  NOT NULL DEFAULT current_timestamp
);

CREATE INDEX idx_1_h_search ON h_search (m_user_id);

CREATE TABLE s_monthly_video_watch
(
    s_monthly_video_watch_id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

    m_user_id                bigint        NOT NULL,
    ai_summary_title         VARCHAR(128)  NOT NULL,
    ai_summary_description   VARCHAR(4096) NOT NULL,
    ai_model                 VARCHAR(128)  NOT NULL,
    generated_at             timestamptz   NOT NULL DEFAULT current_timestamp,
    target_month             date          NOT NULL,

    w_monthly_video_watch_id bigint        NOT NULL,

    created_at               timestamptz   NOT NULL DEFAULT current_timestamp,
    updated_at               timestamptz   NOT NULL DEFAULT current_timestamp,

    public_id                uuid          NOT NULL DEFAULT uuidv7()
);

CREATE UNIQUE INDEX uk_1_s_monthly_video_watch ON s_monthly_video_watch (public_id);
CREATE UNIQUE INDEX uk_2_s_monthly_video_watch ON s_monthly_video_watch (m_user_id, target_month);

CREATE TABLE w_monthly_video_watch
(
    w_monthly_video_watch_id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

    m_user_id                bigint       NOT NULL,
    batch_status_code        int          NOT NULL DEFAULT 1,
    ai_model                 VARCHAR(128) NOT NULL,
    started_at               timestamptz  NOT NULL DEFAULT current_timestamp,
    finished_at              timestamptz  NOT NULL DEFAULT '9999-12-31 23:59:59UTC',

    created_at               timestamptz  NOT NULL DEFAULT current_timestamp,
    updated_at               timestamptz  NOT NULL DEFAULT current_timestamp
);

CREATE TABLE m_playlist
(
    m_playlist_id                       bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

    m_user_id                           bigint       NOT NULL,
    visibility_code                     INTEGER      NOT NULL DEFAULT 0,
    playlist_title                      VARCHAR(128) NOT NULL,
    playlist_code                       INTEGER      NOT NULL DEFAULT 0,

    video_count                         INTEGER      NOT NULL DEFAULT 0,
    playlist_total_video_length_seconds INTEGER      NOT NULL DEFAULT 0,

    created_at                          timestamptz  NOT NULL DEFAULT current_timestamp,
    updated_at                          timestamptz  NOT NULL DEFAULT current_timestamp,

    public_id                           uuid         NOT NULL DEFAULT uuidv7()
);

CREATE UNIQUE INDEX uk_1_m_playlist ON m_playlist (public_id);
CREATE INDEX idx_1_m_playlist ON m_playlist (m_user_id, visibility_code, playlist_code);

CREATE TABLE m_playlist_video
(
    m_playlist_video_id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

    m_playlist_id       bigint      NOT NULL,
    m_video_id          bigint      NOT NULL,

    playlist_position   bigint      NOT NULL,

    created_at          timestamptz NOT NULL DEFAULT current_timestamp,
    updated_at          timestamptz NOT NULL DEFAULT current_timestamp
);

CREATE UNIQUE INDEX uk_1_m_playlist_video ON m_playlist_video (m_playlist_id, m_video_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE m_playlist_video;

DROP TABLE m_playlist;

DROP TABLE w_monthly_video_watch;

DROP TABLE s_monthly_video_watch;

DROP TABLE h_search;

DROP TABLE t_video_watch;

DROP TABLE m_comment;

DROP TABLE m_video;

DROP TABLE m_channel;

DROP TABLE m_user_subscribing_channel;

DROP TABLE m_user_screen_time_range;

DROP TABLE h_user;

DROP TABLE m_refresh_token;

DROP TABLE m_user_authorization;

DROP TABLE m_user;
-- +goose StatementEnd

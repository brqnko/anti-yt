package youtube_d

import (
	"net/url"
	"strings"

	"github.com/brqnko/anti-yt/backend/internal/core"
)

var (
	ErrInvalidChannelHandle = core.NewDomainError("channel.invalid_channel_handle", "invalid channel handle")
	ErrInvalidYouTubeURL    = core.NewDomainError("channel.invalid_youtube_url", "invalid youtube url or unsupported format")
	ErrInvalidPlaylistID    = core.NewDomainError("playlist.invalid_playlist_id", "invalid playlist id or unsupported format")
	ErrInvalidVideoID       = core.NewDomainError("video.invalid_video_id", "invalid video id or unsupported format")
)

func ExtractChannelIDOrHandle(channelText string) (string, error) {
	if strings.HasPrefix(channelText, "@") {
		if len([]rune(channelText)) <= 3 {
			return "", ErrInvalidChannelHandle
		}

		return channelText, nil
	}

	if strings.HasPrefix(channelText, "UC") {
		if len(channelText) != 24 {
			return "", ErrInvalidChannelID
		}

		return channelText, nil
	}

	if !strings.HasPrefix(channelText, "http://") && !strings.HasPrefix(channelText, "https://") {
		channelText = "https://" + channelText
	}
	u, err := url.Parse(channelText)
	if err != nil {
		return "", err
	}

	if !strings.HasSuffix(u.Host, "youtube.com") && u.Host != "youtu.be" {
		return "", ErrInvalidYouTubeURL
	}

	path := strings.Trim(u.Path, "/")
	parts := strings.Split(path, "/")

	if len(parts) == 0 || parts[0] == "" {
		return "", ErrInvalidYouTubeURL
	}

	if strings.HasPrefix(parts[0], "@") {
		return parts[0], nil
	}

	if parts[0] == "channel" && len(parts) > 1 {
		id := parts[1]
		if strings.HasPrefix(id, "UC") && len(id) == 24 {
			return id, nil
		}
	}

	return "", ErrInvalidYouTubeURL
}

// ExtractVideoID はYouTubeの動画URLまたは生の動画IDからYouTube動画IDを抽出する
func ExtractVideoID(videoText string) (string, error) {
	videoText = strings.TrimSpace(videoText)

	// 生の動画ID（11文字の英数字+ハイフン+アンダースコア）
	if len(videoText) == 11 && !strings.Contains(videoText, "/") && !strings.Contains(videoText, ".") {
		return videoText, nil
	}

	if !strings.HasPrefix(videoText, "http://") && !strings.HasPrefix(videoText, "https://") {
		videoText = "https://" + videoText
	}
	u, err := url.Parse(videoText)
	if err != nil {
		return "", ErrInvalidVideoID
	}

	if !strings.HasSuffix(u.Host, "youtube.com") && u.Host != "youtu.be" {
		return "", ErrInvalidVideoID
	}

	// youtu.be/VIDEO_ID
	if u.Host == "youtu.be" {
		path := strings.Trim(u.Path, "/")
		if path == "" {
			return "", ErrInvalidVideoID
		}
		parts := strings.Split(path, "/")
		return parts[0], nil
	}

	// youtube.com/watch?v=VIDEO_ID
	if v := u.Query().Get("v"); v != "" {
		return v, nil
	}

	path := strings.Trim(u.Path, "/")
	parts := strings.Split(path, "/")

	// youtube.com/shorts/VIDEO_ID, youtube.com/embed/VIDEO_ID, youtube.com/v/VIDEO_ID
	if len(parts) >= 2 && (parts[0] == "shorts" || parts[0] == "embed" || parts[0] == "v" || parts[0] == "live") {
		return parts[1], nil
	}

	return "", ErrInvalidVideoID
}

// extractPlaylistID はYouTubeのプレイリストURLまたは生のプレイリストIDからプレイリストIDを抽出する
// NOTE: UU, PL, OL以外はAPIで取得できないので無視する.
// UUはチャンネルのアップロード動画
// PLはプレイリスt
// OLは公式のプレイリスト
func ExtractPlaylistID(playlistText string) (string, error) {
	if strings.HasPrefix(playlistText, "PL") || strings.HasPrefix(playlistText, "UU") || strings.HasPrefix(playlistText, "OL") {
		return playlistText, nil
	}

	if !strings.HasPrefix(playlistText, "http://") && !strings.HasPrefix(playlistText, "https://") {
		playlistText = "https://" + playlistText
	}
	u, err := url.Parse(playlistText)
	if err != nil {
		return "", ErrInvalidPlaylistID
	}

	if u.Host != "youtube.com" && !strings.HasSuffix(u.Host, ".youtube.com") && u.Host != "youtu.be" {
		return "", ErrInvalidPlaylistID
	}

	listID := u.Query().Get("list")
	if listID == "" {
		return "", ErrInvalidPlaylistID
	}

	return listID, nil
}

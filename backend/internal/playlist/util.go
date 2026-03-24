package playlist

import (
	"net/url"
	"strings"
)

// extractPlaylistID はYouTubeのプレイリストURLまたは生のプレイリストIDからプレイリストIDを抽出する
// NOTE: UU, PL, OL以外はAPIで取得できないので無視する.
// UUはチャンネルのアップロード動画
// PLはプレイリスt
// OLは公式のプレイリスト
func extractPlaylistID(playlistText string) (string, error) {
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

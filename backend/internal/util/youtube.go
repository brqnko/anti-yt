package util

import (
	"errors"
	"net/url"
	"strings"
)

var (
	ErrInvalidYouTubeURL    = errors.New("invalid youtube url or unsupported format")
	ErrInvalidChannelID     = errors.New("invalid channel id")
	ErrInvalidChannelHandle = errors.New("invalid channel handle")
)

func ExtractChannelIdOrHandle(channelText string) (string, error) {
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

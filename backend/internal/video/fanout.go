package video

import (
	"context"

	"github.com/brqnko/anti-yt/backend/internal/core/database_d"
	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
	"github.com/brqnko/anti-yt/backend/internal/util"
)

func FanOut(ctx context.Context, q sqlc.Querier, feedRepo database_d.FeedRepository, v *Video) (err error) {
	defer util.Wrap(&err, "video.FanOut(videoID=%s)", v.ID)

	subscribers, err := q.ListSubscribersByChannelPublicID(ctx, v.ChannelID)
	if err != nil {
		return err
	}

	if err := feedRepo.Push(ctx, subscribers, v.ID); err != nil {
		return err
	}
	return nil
}

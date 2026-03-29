package v1

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/brqnko/anti-yt/backend/internal/core/handler/hutil"
)

func (h *APIHandler) GetPlaylistsRecent(ctx context.Context, request GetPlaylistsRecentRequestObject) (GetPlaylistsRecentResponseObject, error) {
	userID, err := hutil.UserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	playlists, err := h.playlistService.GetRecentPlaylists(ctx, userID)
	if err != nil {
		return nil, err
	}

	loc := hutil.TimezoneFromContext(ctx)
	items := make([]struct {
		PlaylistId           uuid.UUID `json:"playlist_id"`
		PlaylistRegisteredAt time.Time `json:"playlist_registered_at"`
		PlaylistTitle        string    `json:"playlist_title"`
		PlaylistVideoCount   int       `json:"playlist_video_count"`
		TopVideoThumbnailUrl *string   `json:"top_video_thumbnail_url,omitempty"`
	}, len(playlists))

	for i, pl := range playlists {
		items[i].PlaylistId = pl.PlaylistId
		items[i].PlaylistTitle = pl.PlaylistTitle
		items[i].PlaylistVideoCount = pl.PlaylistVideoCount
		items[i].PlaylistRegisteredAt = pl.PlaylistRegisteredAt.In(loc)
		items[i].TopVideoThumbnailUrl = pl.TopVideoThumbnailUrl
	}

	return GetPlaylistsRecent200JSONResponse{
		ItemCount: len(items),
		Items:     items,
	}, nil
}

func (h *APIHandler) GetPlaylists(ctx context.Context, request GetPlaylistsRequestObject) (GetPlaylistsResponseObject, error) {
	userID, err := hutil.UserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	playlists, hasNext, err := h.playlistService.GetPlaylists(ctx, userID, request.Params.Cursor, int32(request.Params.Limit))
	if err != nil {
		return nil, err
	}

	resp := GetPlaylists200JSONResponse{
		HasNext:   hasNext,
		ItemCount: len(playlists),
		Items: make([]struct {
			PlaylistDescription  string             `json:"playlist_description"`
			PlaylistId           uuid.UUID          `json:"playlist_id"`
			PlaylistRegisteredAt time.Time          `json:"playlist_registered_at"`
			PlaylistTitle        string             `json:"playlist_title"`
			PlaylistType         PlaylistType       `json:"playlist_type"`
			PlaylistUpdatedAt    time.Time          `json:"playlist_updated_at"`
			PlaylistVideoCount   int                `json:"playlist_video_count"`
			PlaylistVisibility   PlaylistVisibility `json:"playlist_visibility"`
			TopVideoThumbnailUrl *string            `json:"top_video_thumbnail_url,omitempty"`
		}, len(playlists)),
	}

	loc := hutil.TimezoneFromContext(ctx)
	for i, pl := range playlists {
		resp.Items[i].PlaylistId = pl.PlaylistId
		resp.Items[i].PlaylistTitle = pl.PlaylistTitle
		resp.Items[i].PlaylistDescription = pl.PlaylistDescription
		resp.Items[i].PlaylistType = PlaylistType(pl.PlaylistType)
		resp.Items[i].PlaylistVisibility = PlaylistVisibility(pl.PlaylistVisibility)
		resp.Items[i].PlaylistVideoCount = pl.PlaylistVideoCount
		resp.Items[i].PlaylistRegisteredAt = pl.PlaylistRegisteredAt.In(loc)
		resp.Items[i].PlaylistUpdatedAt = pl.PlaylistUpdatedAt.In(loc)
		resp.Items[i].TopVideoThumbnailUrl = pl.TopVideoThumbnailUrl
	}

	return resp, nil
}

func (h *APIHandler) PostPlaylists(ctx context.Context, request PostPlaylistsRequestObject) (PostPlaylistsResponseObject, error) {
	userID, err := hutil.UserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	created, err := h.playlistService.CreatePlaylist(
		ctx,
		userID,
		request.Body.PlaylistTitle,
		request.Body.PlaylistDescription,
		string(request.Body.PlaylistVisibility),
		string(request.Body.PlaylistType),
		request.Body.BasePlaylistUrl,
	)
	if err != nil {
		return nil, err
	}

	return PostPlaylists201JSONResponse{
		PlaylistId:           created.ID,
		PlaylistType:         PlaylistType(created.PlaylistCode.String()),
		PlaylistVisibility:   PlaylistVisibility(created.VisibilityCode.String()),
		PlaylistTitle:        string(created.Title),
		PlaylistDescription:  string(created.Description),
		PlaylistVideoCount:   created.VideoCount,
		PlaylistRegisteredAt: created.RegisteredAt.In(hutil.TimezoneFromContext(ctx)),
		// TODO: updated_atは物理カラムとして存在するがドメインには含めていない。
		// QueryServiceではSQLから直接updated_atを取得して誤魔化しているが、ここではドメイン経由のためRegisteredAtで代用している。
		PlaylistUpdatedAt: created.RegisteredAt.In(hutil.TimezoneFromContext(ctx)),
	}, nil
}

func (h *APIHandler) DeletePlaylistsPlaylistId(ctx context.Context, request DeletePlaylistsPlaylistIdRequestObject) (DeletePlaylistsPlaylistIdResponseObject, error) {
	userID, err := hutil.UserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if err := h.playlistService.DeletePlaylist(ctx, userID, request.PlaylistId); err != nil {
		return nil, err
	}

	return DeletePlaylistsPlaylistId204Response{}, nil
}

func (h *APIHandler) GetPlaylistsPlaylistId(ctx context.Context, request GetPlaylistsPlaylistIdRequestObject) (GetPlaylistsPlaylistIdResponseObject, error) {
	userID, err := hutil.UserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	pl, err := h.playlistService.GetPlaylistDetail(ctx, userID, request.PlaylistId)
	if err != nil {
		return nil, err
	}

	return GetPlaylistsPlaylistId200JSONResponse{
		PlaylistId:           pl.PlaylistId,
		PlaylistTitle:        pl.PlaylistTitle,
		PlaylistDescription:  pl.PlaylistDescription,
		PlaylistType:         PlaylistType(pl.PlaylistType),
		PlaylistVisibility:   PlaylistVisibility(pl.PlaylistVisibility),
		PlaylistVideoCount:   pl.PlaylistVideoCount,
		PlaylistRegisteredAt: pl.PlaylistRegisteredAt.In(hutil.TimezoneFromContext(ctx)),
		PlaylistUpdatedAt:    pl.PlaylistUpdatedAt.In(hutil.TimezoneFromContext(ctx)),
		TopVideoThumbnailUrl: pl.TopVideoThumbnailUrl,
	}, nil
}

func (h *APIHandler) GetPlaylistsPlaylistIdVideos(ctx context.Context, request GetPlaylistsPlaylistIdVideosRequestObject) (GetPlaylistsPlaylistIdVideosResponseObject, error) {
	userID, err := hutil.UserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	videos, hasNext, err := h.playlistService.GetPlaylistItems(ctx, userID, request.PlaylistId, request.Params.Cursor, int32(request.Params.Limit))
	if err != nil {
		return nil, err
	}

	resp := GetPlaylistsPlaylistIdVideos200JSONResponse{
		HasNext:   hasNext,
		ItemCount: len(videos),
		Items: make([]struct {
			ChannelId                  uuid.UUID `json:"channel_id"`
			ExternalChannelDisplayName string    `json:"external_channel_display_name"`
			ExternalChannelIconUrl     string    `json:"external_channel_icon_url"`
			ExternalVideoCreatedAt     time.Time `json:"external_video_created_at"`
			ExternalVideoLengthSeconds int       `json:"external_video_length_seconds"`
			ExternalVideoThumbnailUrl  string    `json:"external_video_thumbnail_url"`
			ExternalVideoTitle         string    `json:"external_video_title"`
			LastWatchSeconds           *int      `json:"last_watch_seconds,omitempty"`
			VideoId                    uuid.UUID `json:"video_id"`
		}, len(videos)),
	}

	loc := hutil.TimezoneFromContext(ctx)
	for i, v := range videos {
		resp.Items[i].VideoId = v.VideoId
		resp.Items[i].ChannelId = v.ChannelId
		resp.Items[i].ExternalVideoThumbnailUrl = v.ExternalVideoThumbnailUrl
		resp.Items[i].ExternalVideoTitle = v.ExternalVideoTitle
		resp.Items[i].ExternalVideoCreatedAt = v.ExternalVideoCreatedAt.In(loc)
		resp.Items[i].ExternalVideoLengthSeconds = v.ExternalVideoLengthSeconds
		resp.Items[i].ExternalChannelIconUrl = v.ExternalChannelIconUrl
		resp.Items[i].ExternalChannelDisplayName = v.ExternalChannelDisplayName
		resp.Items[i].LastWatchSeconds = v.LastWatchSeconds
	}

	return resp, nil
}

func (h *APIHandler) PatchPlaylistsPlaylistId(ctx context.Context, request PatchPlaylistsPlaylistIdRequestObject) (PatchPlaylistsPlaylistIdResponseObject, error) {
	userID, err := hutil.UserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	updated, err := h.playlistService.UpdatePlaylist(ctx, userID, request.PlaylistId, request.Body.PlaylistTitle, request.Body.PlaylistDescription)
	if err != nil {
		return nil, err
	}

	return PatchPlaylistsPlaylistId200JSONResponse{
		PlaylistId:          updated.ID,
		PlaylistTitle:       string(updated.Title),
		PlaylistDescription: string(updated.Description),
	}, nil
}

func (h *APIHandler) DeletePlaylistsPlaylistIdVideos(ctx context.Context, request DeletePlaylistsPlaylistIdVideosRequestObject) (DeletePlaylistsPlaylistIdVideosResponseObject, error) {
	userID, err := hutil.UserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if err := h.playlistService.RemoveVideoFromPlaylist(ctx, userID, request.PlaylistId, request.Params.VideoId); err != nil {
		return nil, err
	}

	return DeletePlaylistsPlaylistIdVideos204Response{}, nil
}

func (h *APIHandler) PostPlaylistsPlaylistIdVideos(ctx context.Context, request PostPlaylistsPlaylistIdVideosRequestObject) (PostPlaylistsPlaylistIdVideosResponseObject, error) {
	userID, err := hutil.UserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	var videoID uuid.UUID

	switch {
	case request.Body.VideoId != nil:
		videoID = *request.Body.VideoId
		if err := h.playlistService.InsertVideoIntoPlaylist(ctx, userID, request.PlaylistId, videoID); err != nil {
			return nil, err
		}
	case request.Body.ExternalVideoText != nil:
		resolved, err := h.playlistService.FetchAndInsertVideoIntoPlaylist(ctx, userID, request.PlaylistId, *request.Body.ExternalVideoText)
		if err != nil {
			return nil, err
		}
		videoID = resolved
	default:
		return PostPlaylistsPlaylistIdVideos400JSONResponse{
			BadRequestJSONResponse{
				Title:  "bad_request",
				Detail: "video_id or external_video_text is required",
			},
		}, nil
	}

	return PostPlaylistsPlaylistIdVideos201JSONResponse{
		PlaylistId: request.PlaylistId,
		VideoId:    videoID,
		InsertedAt: time.Now().In(hutil.TimezoneFromContext(ctx)),
	}, nil
}

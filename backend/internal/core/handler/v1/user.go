package v1

import (
	"context"
	"net/http"

	"github.com/brqnko/anti-yt/backend/internal/core/handler/hutil"
	userdomain "github.com/brqnko/anti-yt/backend/internal/user"
	"github.com/brqnko/anti-yt/backend/internal/util"
)

func (h *APIHandler) DeleteUsersMe(ctx context.Context, request DeleteUsersMeRequestObject) (DeleteUsersMeResponseObject, error) {
	userID, err := hutil.UserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if err := h.userService.RemoveUser(ctx, userID); err != nil {
		return nil, err
	}

	return DeleteUsersMe204Response{}, nil
}

func (h *APIHandler) GetUsersMeStatus(ctx context.Context, request GetUsersMeStatusRequestObject) (GetUsersMeStatusResponseObject, error) {
	userID, err := hutil.UserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	view, err := h.userService.GetUserStatus(ctx, userID)
	if err != nil {
		return nil, err
	}

	loc := hutil.TimezoneFromContext(ctx)
	return GetUsersMeStatus200JSONResponse{
		DailyScreenSeconds: view.DailyScreenSeconds,
		DisplayName:        view.DisplayName,
		Id:                 util.Base64UUID(view.UserID),
		JoinedAt:           view.JoinedAt.In(loc),
		LanguageCode:       view.LanguageCode,
		ScreenTime:         domainRangesToSlots(view.ScreenTimeLimitRange.ToLocalRanges(loc)),
	}, nil
}

func (h *APIHandler) PatchUsersMeStatus(ctx context.Context, request PatchUsersMeStatusRequestObject) (PatchUsersMeStatusResponseObject, error) {
	var screenTimeDto *[]struct{ Start, End int }
	if request.Body.ScreenTime != nil {
		converted, err := screenTimeSlotsToDto(*request.Body.ScreenTime)
		if err != nil {
			return nil, err
		}
		screenTimeDto = &converted
	}
	userID, err := hutil.UserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	loc := hutil.TimezoneFromContext(ctx)
	u, rangeSet, err := h.userService.EditUser(ctx, userID, request.Body.DisplayName, request.Body.LanguageCode, request.Body.DailyScreenSeconds, screenTimeDto, loc)
	if err != nil {
		return nil, err
	}

	return PatchUsersMeStatus200JSONResponse{
		DailyScreenSeconds: u.ScreenTimeLimit.ToIntPtr(),
		DisplayName:        u.DisplayName.String(),
		Id:                 util.Base64UUID(u.ID),
		JoinedAt:           u.JoinedAt.In(loc),
		LanguageCode:       u.LanguageCode.String(),
		ScreenTime:         domainRangesToSlots(rangeSet.ToLocalRanges(loc)),
	}, nil
}

func (h *APIHandler) PostUsersMe(ctx context.Context, request PostUsersMeRequestObject) (PostUsersMeResponseObject, error) {
	screenTimeDto, err := screenTimeSlotsToDto(request.Body.ScreenTime)
	if err != nil {
		return nil, err
	}
	accessToken, ok := hutil.AccessTokenFromContext(ctx)
	if !ok {
		return nil, hutil.ErrUserIDNotFoundInContext
	}

	loc := hutil.TimezoneFromContext(ctx)
	u, rangeSet, newAccessToken, accessTokenExpiresAt, err := h.userService.CreateNewUser(
		ctx,
		accessToken,
		request.Body.DailyScreenSeconds,
		screenTimeDto,
		request.Body.DisplayName,
		request.Body.LanguageCode,
		loc,
	)
	if err != nil {
		return nil, err
	}

	// RegisterTokenからUserAccessTokenに切り替え
	hutil.AddResponseCookie(ctx, (new(http.Cookie{
		Name:     "access_token",
		Value:    newAccessToken,
		Path:     "/",
		Expires:  accessTokenExpiresAt,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})).String())

	return PostUsersMe201JSONResponse{
		DailyScreenSeconds: u.ScreenTimeLimit.ToIntPtr(),
		DisplayName:        u.DisplayName.String(),
		Id:                 util.Base64UUID(u.ID),
		JoinedAt:           u.JoinedAt.In(loc),
		LanguageCode:       u.LanguageCode.String(),
		ScreenTime:         domainRangesToSlots(rangeSet.ToLocalRanges(loc)),
	}, nil
}

func screenTimeSlotsToDto(slots []ScreenTimeSlot) ([]struct{ Start, End int }, error) {
	result := make([]struct{ Start, End int }, len(slots))
	for i, s := range slots {
		start, err := util.HHmmToInt(s.StartTime)
		if err != nil {
			return nil, err
		}
		end, err := util.HHmmToInt(s.EndTime)
		if err != nil {
			return nil, err
		}
		result[i] = struct{ Start, End int }{start * 60, end * 60}
	}
	return result, nil
}

func domainRangesToSlots(ranges []userdomain.DailyScreenTimeLimitRange) []ScreenTimeSlot {
	slots := make([]ScreenTimeSlot, len(ranges))
	for i, r := range ranges {
		startStr, _ := util.IntToHHmm(r.StartTimeSeconds / 60)
		endStr, _ := util.IntToHHmm(r.EndTimeSeconds / 60)
		slots[i] = ScreenTimeSlot{StartTime: startStr, EndTime: endStr}
	}
	return slots
}

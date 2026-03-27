package v1

import (
	"context"
	"net/http"
	"sort"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core/handler/hutil"
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

	user, err := h.userService.GetUserStatus(ctx, userID)
	if err != nil {
		return nil, err
	}

	loc := hutil.TimezoneFromContext(ctx)
	_, offsetSec := time.Now().In(loc).Zone()
	const daySec = 24 * 60 * 60

	// UTC HH:mm → ローカル秒に変換
	type secRange struct{ start, end int }
	localRanges := make([]secRange, 0, len(user.ScreenTimeLimitRange))
	for _, v := range user.ScreenTimeLimitRange {
		startMin, err := util.HHmmToInt(v.StartTime)
		if err != nil {
			return nil, err
		}
		endMin, err := util.HHmmToInt(v.EndTime)
		if err != nil {
			return nil, err
		}
		startSec := startMin * 60
		endSec := endMin * 60
		// [0, 86400] は全日を表すため、タイムゾーン変換せずそのまま返す
		// (endSec % daySec == 0 になりオフセットが加算されると開始と同じ値になるため)
		if startSec == 0 && endSec == daySec {
			localRanges = append(localRanges, secRange{start: 0, end: daySec})
			continue
		}
		localRanges = append(localRanges, secRange{
			start: ((startSec + offsetSec) % daySec + daySec) % daySec,
			end:   ((endSec + offsetSec) % daySec + daySec) % daySec,
		})
	}

	// ソートしてマージ
	sort.Slice(localRanges, func(i, j int) bool { return localRanges[i].start < localRanges[j].start })
	var merged []secRange
	for _, r := range localRanges {
		if len(merged) > 0 && r.start <= merged[len(merged)-1].end+1 {
			if r.end > merged[len(merged)-1].end {
				merged[len(merged)-1].end = r.end
			}
		} else {
			merged = append(merged, r)
		}
	}

	screenTime := make([]ScreenTimeSlot, len(merged))
	for i, r := range merged {
		startStr, err := util.IntToHHmm(r.start / 60)
		if err != nil {
			return nil, err
		}
		endStr, err := util.IntToHHmm(r.end / 60)
		if err != nil {
			return nil, err
		}
		screenTime[i] = ScreenTimeSlot{
			StartTime: startStr,
			EndTime:   endStr,
		}
	}

	return GetUsersMeStatus200JSONResponse{
		DailyScreenSeconds: user.DailyScreenSeconds,
		DisplayName:        user.DisplayName,
		Id:                 user.UserID,
		JoinedAt:           user.JoinedAt.In(loc),
		LanguageCode:       user.LanguageCode,
		ScreenTime:         screenTime,
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

	u, err := h.userService.EditUser(ctx, userID, request.Body.DisplayName, request.Body.LanguageCode, request.Body.DailyScreenSeconds, screenTimeDto, hutil.TimezoneFromContext(ctx))
	if err != nil {
		return nil, err
	}

	return PatchUsersMeStatus200JSONResponse{
		DailyScreenSeconds: u.ScreenTimeLimit.ToIntPtr(),
		DisplayName:        u.DisplayName.String(),
		Id:                 u.ID,
		JoinedAt:           u.JoinedAt.In(hutil.TimezoneFromContext(ctx)),
		LanguageCode:       u.LanguageCode.String(),
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

	u, newAccessToken, accessTokenExpiresAt, err := h.userService.CreateNewUser(ctx, accessToken, request.Body.DailyScreenSeconds, screenTimeDto, request.Body.DisplayName, request.Body.LanguageCode, hutil.TimezoneFromContext(ctx))
	if err != nil {
		return nil, err
	}

	// RegisterTokenからUserAccessTokenに切り替え
	hutil.AddResponseCookie(ctx, (&http.Cookie{
		Name:     "access_token",
		Value:    newAccessToken,
		Path:     "/",
		Expires:  accessTokenExpiresAt,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}).String())

	return PostUsersMe201JSONResponse{
		DailyScreenSeconds: u.ScreenTimeLimit.ToIntPtr(),
		DisplayName:        u.DisplayName.String(),
		Id:                 u.ID,
		JoinedAt:           u.JoinedAt.In(hutil.TimezoneFromContext(ctx)),
		LanguageCode:       u.LanguageCode.String(),
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

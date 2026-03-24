package v1

import (
	"context"
	"errors"
	"net/http"

	"github.com/brqnko/anti-yt/backend/internal/core/handler/hutil"
	"github.com/brqnko/anti-yt/backend/internal/user"
	"github.com/brqnko/anti-yt/backend/internal/util"
)

func (h *APIHandler) DeleteUsersMe(ctx context.Context, request DeleteUsersMeRequestObject) (DeleteUsersMeResponseObject, error) {
	userID, err := hutil.UserIDFromContext(ctx)
	if err != nil {
		hutil.LogError(ctx, err)
		return DeleteUsersMe500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Detail: internalErrorDetail,
				Title:  internalErrorTitle,
			},
		}, nil
	}

	if err := h.userService.RemoveUser(ctx, userID); err != nil {
		hutil.LogError(ctx, err)
		return DeleteUsersMe500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Detail: internalErrorDetail,
				Title:  internalErrorTitle,
			},
		}, nil
	}

	return DeleteUsersMe204Response{}, nil
}

func (h *APIHandler) GetUsersMeStatus(ctx context.Context, request GetUsersMeStatusRequestObject) (GetUsersMeStatusResponseObject, error) {
	userID, err := hutil.UserIDFromContext(ctx)
	if err != nil {
		hutil.LogError(ctx, err)
		return GetUsersMeStatus500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Detail: internalErrorDetail,
				Title:  internalErrorTitle,
			},
		}, nil
	}

	user, err := h.userService.GetUserStatus(ctx, userID)
	if err != nil {
		hutil.LogError(ctx, err)
		return GetUsersMeStatus500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Detail: internalErrorDetail,
				Title:  internalErrorTitle,
			},
		}, nil
	}

	screenTime := make([]ScreenTimeSlot, len(user.ScreenTimeLimitRange))
	for i, v := range user.ScreenTimeLimitRange {
		screenTime[i] = ScreenTimeSlot{
			EndTime:   v.EndTime,
			Id:        v.ID,
			StartTime: v.StartTime,
		}
	}

	return GetUsersMeStatus200JSONResponse{
		DailyScreenSeconds: user.DailyScreenSeconds,
		DisplayName:        user.DisplayName,
		Id:                 user.UserID,
		JoinedAt:           user.JoinedAt,
		LanguageCode:       user.LanguageCode,
		ScreenTime:         screenTime,
	}, nil
}

func (h *APIHandler) PatchUsersMeStatus(ctx context.Context, request PatchUsersMeStatusRequestObject) (PatchUsersMeStatusResponseObject, error) {
	var screenTimeDto *[]struct{ Start, End int }
	if request.Body.ScreenTime != nil {
		converted, err := screenTimeSlotsToDto(*request.Body.ScreenTime)
		if err != nil {
			hutil.LogError(ctx, err)
			return PatchUsersMeStatus500JSONResponse{
				InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
					Detail: internalErrorDetail,
					Title:  internalErrorTitle,
				},
			}, nil
		}
		screenTimeDto = &converted
	}
	userID, err := hutil.UserIDFromContext(ctx)
	if err != nil {
		hutil.LogError(ctx, err)
		return PatchUsersMeStatus500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Detail: internalErrorDetail,
				Title:  internalErrorTitle,
			},
		}, nil
	}

	u, err := h.userService.EditUser(ctx, userID, request.Body.DisplayName, request.Body.LanguageCode, request.Body.DailyScreenSeconds, screenTimeDto)
	if err != nil {
		if br := userDomainErrToBadRequest(err); br != nil {
			return PatchUsersMeStatus400JSONResponse{BadRequestJSONResponse: *br}, nil
		}
		hutil.LogError(ctx, err)
		return PatchUsersMeStatus500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Detail: internalErrorDetail,
				Title:  internalErrorTitle,
			},
		}, nil
	}

	return PatchUsersMeStatus200JSONResponse{
		DailyScreenSeconds: u.ScreenTimeLimit.ToIntPtr(),
		DisplayName:        u.DisplayName.String(),
		Id:                 u.ID,
		JoinedAt:           u.JoinedAt,
		LanguageCode:       u.LanguageCode.String(),
	}, nil
}

func (h *APIHandler) PostUsersMe(ctx context.Context, request PostUsersMeRequestObject) (PostUsersMeResponseObject, error) {
	screenTimeDto, err := screenTimeSlotsToDto(request.Body.ScreenTime)
	if err != nil {
		hutil.LogError(ctx, err)
		return PostUsersMe500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Title:  internalErrorTitle,
				Detail: internalErrorDetail,
			},
		}, nil
	}
	accessToken, ok := hutil.AccessTokenFromContext(ctx)
	if !ok {
		return PostUsersMe500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Detail: internalErrorDetail,
				Title:  internalErrorTitle,
			},
		}, nil
	}

	u, newAccessToken, accessTokenExpiresAt, err := h.userService.CreateNewUser(ctx, accessToken, request.Body.DailyScreenSeconds, screenTimeDto, request.Body.DisplayName, request.Body.LanguageCode)
	if err != nil {
		if br := userDomainErrToBadRequest(err); br != nil {
			return PostUsersMe400JSONResponse{BadRequestJSONResponse: *br}, nil
		}
		hutil.LogError(ctx, err)
		return PostUsersMe500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Title:  internalErrorTitle,
				Detail: internalErrorDetail,
			},
		}, nil
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
		JoinedAt:           u.JoinedAt,
		LanguageCode:       u.LanguageCode.String(),
	}, nil
}

func userDomainErrToBadRequest(err error) *BadRequestJSONResponse {
	var detail string
	switch {
	case errors.Is(err, user.ErrDailyScreenTimeLimitOutOfRange), errors.Is(err, user.ErrDailyScreenTimeOutOfRange):
		detail = "daily screen time limit is out of range"
	case errors.Is(err, user.ErrDailyScreenTimeLimitRangeOrder):
		detail = "screen time range start must be before end"
	case errors.Is(err, user.ErrDisplayNameTooLong):
		detail = "display name is too long"
	case errors.Is(err, user.ErrDisplayNameTooShort):
		detail = "display name is too short"
	case errors.Is(err, user.ErrLanguageCodeNotSupported):
		detail = "language code is not supported"
	default:
		return nil
	}
	return &BadRequestJSONResponse{Detail: detail, Title: detail}
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
		result[i] = struct{ Start, End int }{start, end}
	}
	return result, nil
}

package v1

import (
	"context"
	"errors"
	"log/slog"

	"github.com/brqnko/anti-yt/backend/internal/user"
	"github.com/brqnko/anti-yt/backend/internal/util"
	"github.com/google/uuid"
)

func (h *APIHandler) DeleteUsersMe(c context.Context, request DeleteUsersMeRequestObject) (DeleteUsersMeResponseObject, error) {
	if err := h.userService.RemoveUser(c); err != nil {
		util.LogError(c, err)
		return DeleteUsersMe500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Detail: internalErrorDetail,
				Title:  internalErrorTitle,
			},
		}, nil
	}

	return DeleteUsersMe204Response{}, nil
}

func (h *APIHandler) GetUsersMeStatus(c context.Context, request GetUsersMeStatusRequestObject) (GetUsersMeStatusResponseObject, error) {
	user, err := h.userService.GetUserStatus(c)
	if err != nil {
		util.LogError(c, err)
		return GetUsersMeStatus500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Detail: internalErrorDetail,
				Title:  internalErrorTitle,
			},
		}, nil
	}

	screenTime, err := dtoToScreenTimeSlots(user.ScreenTimeLimitRange)
	if err != nil {
		util.LogError(c, err)
		return GetUsersMeStatus500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Detail: internalErrorDetail,
				Title:  internalErrorTitle,
			},
		}, nil
	}

	return GetUsersMeStatus200JSONResponse{
		DailyRemainingSeconds: user.RemainingSeconds,
		DailyScreenSeconds:    user.ScreenTimeSeconds,
		DisplayName:           user.DisplayName,
		Id:                    user.UserID,
		JoinedAt:              user.JoinedAt,
		LanguageCode:          &user.LanguageCode,
		ScreenTime:            screenTime,
	}, nil
}

func (h *APIHandler) PatchUsersMeStatus(c context.Context, request PatchUsersMeStatusRequestObject) (PatchUsersMeStatusResponseObject, error) {
	var screenTimeDto *[]struct{ Start, End int }
	if request.Body.ScreenTime != nil {
		converted, err := screenTimeSlotsToDto(*request.Body.ScreenTime)
		if err != nil {
			util.LogError(c, err)
			return PatchUsersMeStatus500JSONResponse{
				InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
					Detail: internalErrorDetail,
					Title:  internalErrorTitle,
				},
			}, nil
		}
		screenTimeDto = &converted
	}
	u, err := h.userService.EditUser(c, request.Body.DisplayName, request.Body.LanguageCode, request.Body.DailyScreenSeconds, screenTimeDto)
	if err != nil {
		if br := userDomainErrToBadRequest(err); br != nil {
			return PatchUsersMeStatus400JSONResponse{BadRequestJSONResponse: *br}, nil
		}
		util.LogError(c, err)
		return PatchUsersMeStatus500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Detail: internalErrorDetail,
				Title:  internalErrorTitle,
			},
		}, nil
	}

	screenTime, err := dtoToScreenTimeSlots(u.ScreenTimeLimitRange)
	if err != nil {
		util.LogError(c, err)
		return PatchUsersMeStatus500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Title:  internalErrorTitle,
				Detail: internalErrorDetail,
			},
		}, nil
	}

	return PatchUsersMeStatus200JSONResponse{
		DailyRemainingSeconds: u.RemainingSeconds,
		DailyScreenSeconds:    u.ScreenTimeSeconds,
		DisplayName:           u.DisplayName,
		Id:                    u.UserID,
		JoinedAt:              u.JoinedAt,
		LanguageCode:          &u.LanguageCode,
		ScreenTime:            screenTime,
	}, nil
}

func (h *APIHandler) PostUsersMe(c context.Context, request PostUsersMeRequestObject) (PostUsersMeResponseObject, error) {
	screenTimeDto, err := screenTimeSlotsToDto(request.Body.ScreenTime)
	if err != nil {
		util.LogError(c, err)
		return PostUsersMe500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Title:  internalErrorTitle,
				Detail: internalErrorDetail,
			},
		}, nil
	}
	u, err := h.userService.CreateNewUser(c, request.Body.DailyScreenSeconds, screenTimeDto, request.Body.DisplayName, request.Body.LanguageCode)
	if err != nil {
		if br := userDomainErrToBadRequest(err); br != nil {
			return PostUsersMe400JSONResponse{BadRequestJSONResponse: *br}, nil
		}
		util.LogError(c, err)
		return PostUsersMe500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Title:  internalErrorTitle,
				Detail: internalErrorDetail,
			},
		}, nil
	}

	screenTime, err := dtoToScreenTimeSlots(u.ScreenTimeLimitRange)
	if err != nil {
		util.LogError(c, err)
		return PostUsersMe500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Title:  internalErrorTitle,
				Detail: internalErrorDetail,
			},
		}, nil
	}

	return PostUsersMe201JSONResponse{
		DailyRemainingSeconds: u.RemainingSeconds,
		DailyScreenSeconds:    u.ScreenTimeSeconds,
		DisplayName:           u.DisplayName,
		Id:                    u.UserID,
		JoinedAt:              u.JoinedAt,
		LanguageCode:          &u.LanguageCode,
		ScreenTime:            screenTime,
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

func dtoToScreenTimeSlots(screenTimeLimitRange []struct {
	ID                       uuid.UUID
	StartSeconds, EndSeconds int
}) ([]ScreenTimeSlot, error) {
	screenTime := make([]ScreenTimeSlot, len(screenTimeLimitRange))
	for i, dto := range screenTimeLimitRange {
		startTime, err := util.IntToHHmm(dto.StartSeconds)
		if err != nil {
			slog.Error("int to hh:mm", "error", err)
			return []ScreenTimeSlot{}, err
		}
		endTime, err := util.IntToHHmm(dto.EndSeconds)
		if err != nil {
			slog.Error("int to hh:mm", "error", err)
			return []ScreenTimeSlot{}, err
		}

		screenTime[i] = ScreenTimeSlot{
			EndTime:   endTime,
			Id:        dto.ID,
			StartTime: startTime,
		}
	}

	return screenTime, nil
}

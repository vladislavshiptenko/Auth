package restoreuser

import (
	resp "auth/internal/lib/api/response"
	"auth/internal/lib/logger/sl"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator"
	"log/slog"
	"net/http"
)

type Request struct {
	UserId int64 `json:"user_id" validate:"required"`
}

type Response struct {
	resp.Response
}

type UserService interface {
	RestoreUser(userId int64) error
}

func New(log *slog.Logger, userService UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.deleteuser.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request
		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to decode request"))

			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			log.Error("invalid request", sl.Err(err))

			render.JSON(w, r, resp.Error("invalid request"))

			return
		}

		err = userService.RestoreUser(req.UserId)

		if err != nil {
			log.Error("failed to restore user", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to restore user"))

			return
		}

		render.JSON(w, r, Response{
			Response: resp.Ok(),
		})
	}
}

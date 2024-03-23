package updateuser

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
	NewFullName string `json:"new_full_name" validate:"required"`
	NewPhone    string `json:"new_phone" validate:"required"`
	NewEmail    string `json:"new_email" validate:"required"`
	UserId      int64  `json:"user_id" validate:"required"`
}

type Response struct {
	resp.Response
}

type UserService interface {
	UpdateUser(newFullName, newPhone, newEmail string, userId int64) error
}

func New(log *slog.Logger, userService UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.updateuser.New"

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

		err = userService.UpdateUser(req.NewFullName, req.NewPhone, req.NewEmail, req.UserId)

		if err != nil {
			log.Error("failed to update user", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to update user"))

			return
		}

		render.JSON(w, r, Response{
			Response: resp.Ok(),
		})
	}
}

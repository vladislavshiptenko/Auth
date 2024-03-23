package restorepassword

import (
	"auth/internal/domain/models"
	resp "auth/internal/lib/api/response"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"time"
)

import (
	"auth/internal/lib/logger/sl"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator"
	"net/http"
)

type Request struct {
	Link        string `json:"link"`
	NewPassword string `json:"new_password"`
}

type Response struct {
	resp.Response
}

type PasswordUpdater interface {
	UpdatePassword(userId int64, newPassword string) error
}

type LinkProvider interface {
	ForgetPasswordInfo(link string) (*models.ForgetPasswordInfo, error)
	DeleteLinkById(id int64) error
}

func New(log *slog.Logger, passwordUpdater PasswordUpdater, linkProvider LinkProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.restorepassword.New"

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

		linkInfo, err := linkProvider.ForgetPasswordInfo(req.Link)
		if err != nil {
			log.Error("user get error", sl.Err(err))

			render.JSON(w, r, resp.Error("invalid request"))

			return
		}

		defer func() {
			err := linkProvider.DeleteLinkById(linkInfo.Id)
			if err != nil {
				sl.Err(err)
			}
		}()

		if time.Now().After(linkInfo.Expiration) {
			log.Error("link is deprecated", sl.Err(err))

			render.JSON(w, r, resp.Error("link is deprecated"))

			return
		}

		passHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			log.Error("failed to generate password hash", sl.Err(err))

			render.JSON(w, r, resp.Error("user register failed"))

			return
		}

		err = passwordUpdater.UpdatePassword(linkInfo.UserId, string(passHash))
		if err != nil {
			log.Error("update password error", sl.Err(err))

			render.JSON(w, r, resp.Error("update password error"))

			return
		}

		render.JSON(w, r, Response{
			Response: resp.Ok(),
		})
	}
}

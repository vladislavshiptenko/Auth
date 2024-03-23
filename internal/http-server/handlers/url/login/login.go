package login

import (
	"auth/internal/domain/models"
	resp "auth/internal/lib/api/response"
	"auth/internal/lib/jwt"
	"auth/internal/lib/logger/sl"
	"auth/internal/storage"
	"errors"
	"fmt"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator"
	"log/slog"
	"net/http"
	"time"
)

type Request struct {
	ContactInfo string `json:"contact_info" validate:"required"`
	Password    string `json:"password" validate:"required"`
}

type Response struct {
	resp.Response
	Token string
}

type UserService interface {
	UserByPhone(email string) (*models.User, error)
	UserByEmail(email string) (*models.User, error)
	UserByContactInfo(contactInfo string) (*models.User, error)
	Authorize(user *models.User, password string) error
}

func New(log *slog.Logger, userService UserService, tokenTtl time.Duration, secretKey string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.authentication.New"

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

		user, err := userService.UserByContactInfo(req.ContactInfo)
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Info("user not found", slog.String("contact_info", req.ContactInfo))

			render.JSON(w, r, resp.Error("user not found"))

			return
		}

		if err != nil {
			log.Error("failed to authentication", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to authentication"))

			return
		}

		if err := userService.Authorize(user, req.Password); err != nil {
			log.Error("invalid credentials", fmt.Errorf(err.Error()))

			render.JSON(w, r, resp.Error("invalid credentials"))

			return
		}

		log.Info("user logged in successfully")

		token, err := jwt.NewToken(*user, secretKey, tokenTtl)
		if err != nil {
			log.Error("failed to generate token", fmt.Errorf(err.Error()))

			render.JSON(w, r, resp.Error("failed to authentication"))

			return
		}

		render.JSON(w, r, Response{
			Response: resp.Ok(),
			Token:    token,
		})
	}
}

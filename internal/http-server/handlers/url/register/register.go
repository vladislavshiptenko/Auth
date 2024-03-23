package register

import (
	resp "auth/internal/lib/api/response"
	"auth/internal/lib/logger/sl"
	"auth/internal/services/auth"
	"auth/internal/storage"
	"errors"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"net/http"
)

type Request struct {
	FullName string `json:"full_name" validate:"required"`
	Password string `json:"password" validate:"required"`
	Phone    string `json:"phone" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	RoleId   string `json:"user_role" validate:"required"`
}

type Response struct {
	resp.Response
}

type RegistrationService interface {
	RegisterUser(fullName, password, phone, email string, userRole string) error
}

func New(log *slog.Logger, registrationService RegistrationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.register.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		log.Info("registering user")

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

		passHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Error("failed to generate password hash", sl.Err(err))

			render.JSON(w, r, resp.Error("user register failed"))

			return
		}

		err = registrationService.RegisterUser(req.FullName, string(passHash), req.Phone, req.Email, req.RoleId)
		if errors.Is(err, storage.ErrUserExist) {
			log.Info("user already exists", slog.String("email", req.Email), slog.String("phone", req.Phone))

			render.JSON(w, r, resp.Error("user with this email or phone already exists"))

			return
		}

		if errors.Is(err, auth.ErrBadPassword) {
			log.Info("bad password", slog.String("password", req.Password))

			render.JSON(w, r, resp.Error("bad password"))

			return
		}

		if err != nil {
			log.Error("failed to add user", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to add user"))

			return
		}

		log.Info("user added")

		render.JSON(w, r, Response{
			Response: resp.Ok(),
		})
	}
}

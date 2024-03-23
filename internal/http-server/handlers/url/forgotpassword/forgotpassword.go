package forgotpassword

import (
	"auth/internal/domain/models"
	resp "auth/internal/lib/api/response"
	"auth/internal/lib/email"
	"auth/internal/lib/logger/sl"
	"auth/internal/storage"
	"errors"
	"fmt"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator"
	"github.com/google/uuid"
	"log/slog"
	"net/http"
	"time"
)

type Request struct {
	ContactInfo string `json:"contact_info" validate:"required"`
}

type Response struct {
	resp.Response
}

type LinkProvider interface {
	SaveLink(link string, linkTtl time.Duration, userId int64) error
}

type UserService interface {
	UserByContactInfo(contactInfo string) (*models.User, error)
}

const (
	emailSubject = "Restore Password"
)

func New(log *slog.Logger, linkProvider LinkProvider, userService UserService, client *http.Client, linkTtl time.Duration, apiKey string, senderName string, senderEmail string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.forgotpassword.New"

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
			log.Info("failed to get user", slog.String("contact_info", req.ContactInfo))

			render.JSON(w, r, resp.Error("failed to get user"))

			return
		}

		link := uuid.New()
		err = linkProvider.SaveLink(link.String(), linkTtl, user.Id)
		if err != nil {
			log.Error("failed to get link", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to get link"))

			return
		}

		emailRequest, err := email.FormSendEmail(apiKey, senderName, senderEmail, user.Email, emailSubject, fmt.Sprintf("http://vacancy/api/auth/restore-password/%s", link.String()))
		response, err := client.Do(emailRequest)

		if err != nil || response.StatusCode != http.StatusOK {
			log.Error("failed to send email", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to send email"))

			return
		}

		render.JSON(w, r, Response{
			Response: resp.Ok(),
		})
	}
}

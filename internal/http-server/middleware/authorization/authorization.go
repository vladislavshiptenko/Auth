package authorization

import (
	"auth/internal/domain/models"
	resp "auth/internal/lib/api/response"
	"auth/internal/lib/logger/sl"
	"bytes"
	"encoding/json"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator"
	"github.com/golang-jwt/jwt/v5"
	"io"
	"log/slog"
	"net/http"
)

const (
	BearerSchema               = "Bearer "
	ParameterAuthorizationName = "Authorization"
)

type Request struct {
	UserId int64 `json:"user_id" validate:"required"`
}

type UserProvider interface {
	UserByUserId(userId int64) (*models.User, error)
}

func New(next http.Handler, log *slog.Logger, secretKey string, userProvider UserProvider, allowedUserRole []models.UserRole) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "middleware.authorization.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		if allowedUserRole == nil || len(allowedUserRole) == 0 {
			log.Error("empty user role")

			render.JSON(w, r, resp.Error("empty user role"))

			return
		}

		tokenWithHeader := r.Header.Get(ParameterAuthorizationName)
		if tokenWithHeader == "" {
			log.Error("token doesn't exist")

			render.JSON(w, r, resp.Error("failed to authentication"))

			return
		}

		tokenString := tokenWithHeader[len(BearerSchema):]
		if tokenString == "" {
			log.Error("token doesn't exist")

			render.JSON(w, r, resp.Error("failed to authentication"))

			return
		}

		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(secretKey), nil
		})

		if err != nil || !token.Valid {
			log.Error("invalid token", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to authentication"))

			return
		}

		var jwtUserId int64
		jwtUserId = int64(claims["user_id"].(float64))

		buf, err := io.ReadAll(r.Body)
		if err != nil {
			log.Error("invalid token", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to authentication"))

			return
		}

		copyBody := io.NopCloser(bytes.NewBuffer(buf))
		r.Body = copyBody

		var req Request
		err = json.Unmarshal(buf, &req)

		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to decode request"))

			return
		}

		if err := validator.New().Struct(req); err != nil {
			log.Error("invalid request", sl.Err(err))

			render.JSON(w, r, resp.Error("invalid request"))

			return
		}

		jwtUser, err := userProvider.UserByUserId(jwtUserId)
		if err != nil || jwtUser == nil {
			log.Error("user from jwt not found", sl.Err(err))

			render.JSON(w, r, resp.Error("invalid authorization"))

			return
		}

		reqUser, err := userProvider.UserByUserId(req.UserId)
		if err != nil || reqUser == nil {
			log.Error("user from jwt not found", sl.Err(err))

			render.JSON(w, r, resp.Error("invalid authorization"))

			return
		}

		if reqUser.Role != models.Admin {
			var necessaryRoleFound bool

			for _, role := range allowedUserRole {
				if role == reqUser.Role {
					necessaryRoleFound = true
					break
				}
			}

			if !necessaryRoleFound || reqUser.Id != jwtUser.Id {
				log.Error("access not allowed")

				render.JSON(w, r, resp.Error("access not allowed"))

				return
			}
		}

		next.ServeHTTP(w, r)
	}
}

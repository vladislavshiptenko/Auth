package auth

import (
	"auth/internal/domain/models"
	"auth/internal/lib/enums"
	"auth/internal/storage"
	"errors"
	"fmt"
	passwordvalidator "github.com/wagslane/go-password-validator"
	"golang.org/x/crypto/bcrypt"
	"time"
)

var (
	ErrEmptyFieldLogin    = errors.New("empty field in authentication")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrBadPassword        = errors.New("bad password")
	EmptyUser             = errors.New("empty user")
	WrongUserRole         = errors.New("wrong user role")
	EmptyEmailErr         = errors.New("email is empty")
	EmptyPhoneErr         = errors.New("phone is empty")
	EmptyNameErr          = errors.New("empty name err")
)

const (
	minEntropyBits = 60
)

type UserRepository interface {
	SaveUser(fullName, password, phone, email string, userRole string) error
	UserByEmail(email string) (*models.User, error)
	UserByPhone(email string) (*models.User, error)
	UpdatePassword(userId int64, newPassword string) error
	UserByUserId(userId int64) (*models.User, error)
	UpdateUser(newFullName, newPhone, newEmail string, userId int64) error
	DeleteUser(userId int64) error
	RestoreUser(userId int64) error
}

type Service struct {
	userRepository UserRepository
	tokenTtl       time.Duration
}

func New(userRepository UserRepository, tokenTtl time.Duration) *Service {
	return &Service{
		userRepository: userRepository,
		tokenTtl:       tokenTtl,
	}
}

func (s *Service) RegisterUser(fullName, password, phone, email string, userRole string) error {
	err := passwordvalidator.Validate(password, minEntropyBits)
	if err != nil {
		return ErrBadPassword
	}

	return s.userRepository.SaveUser(fullName, password, phone, email, userRole)
}

func (s *Service) UpdatePassword(userId int64, newPassword string) error {
	err := passwordvalidator.Validate(newPassword, minEntropyBits)
	if err != nil {
		return ErrBadPassword
	}

	return s.userRepository.UpdatePassword(userId, newPassword)
}

func (s *Service) UserByEmail(email string) (*models.User, error) {
	if email == "" {
		return nil, ErrEmptyFieldLogin
	}

	return s.userRepository.UserByEmail(email)
}

func (s *Service) UserByPhone(phone string) (*models.User, error) {
	if phone == "" {
		return nil, ErrEmptyFieldLogin
	}

	return s.userRepository.UserByPhone(phone)
}

func (s *Service) UserByContactInfo(contactInfo string) (*models.User, error) {
	if contactInfo == "" {
		return nil, ErrEmptyFieldLogin
	}

	user, err := s.UserByPhone(contactInfo)
	fmt.Println(user)
	if err == nil {
		return user, nil
	}
	if !errors.Is(err, storage.ErrUserNotFound) {
		return nil, err
	}

	user, err = s.UserByEmail(contactInfo)
	if err == nil {
		return user, nil
	}
	if !errors.Is(err, storage.ErrUserNotFound) {
		return nil, err
	}

	return nil, storage.ErrUserNotFound
}

func (s *Service) Authorize(user *models.User, password string) error {
	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		return ErrInvalidCredentials
	}

	return nil
}

func (s *Service) UserByUserId(userId int64) (*models.User, error) {
	if userId == 0 {
		return nil, EmptyUser
	}

	user, err := s.userRepository.UserByUserId(userId)
	if err != nil {
		return nil, err
	}

	user.Role = enums.RoleConvertFromString(user.RoleString)
	if user.Role == 0 {
		return nil, WrongUserRole
	}

	return user, nil
}

func (s *Service) UpdateUser(newFullName, newPhone, newEmail string, userId int64) error {
	if newFullName == "" {
		return EmptyNameErr
	}

	if newPhone == "" {
		return EmptyPhoneErr
	}

	if newEmail == "" {
		return EmptyEmailErr
	}

	if userId == 0 {
		return EmptyUser
	}

	return s.userRepository.UpdateUser(newFullName, newPhone, newEmail, userId)
}

func (s *Service) DeleteUser(userId int64) error {
	if userId == 0 {
		return EmptyUser
	}

	return s.userRepository.DeleteUser(userId)
}

func (s *Service) RestoreUser(userId int64) error {
	if userId == 0 {
		return EmptyUser
	}

	return s.userRepository.RestoreUser(userId)
}

package postgres

import (
	"auth/internal/domain/models"
	"auth/internal/storage"
	"database/sql"
	"errors"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/lib/pq"
	"log"
	"time"
)

const (
	UniqueViolationCode = "23505"
)

func (s *Storage) SaveUser(fullName, password, phone, email string, userRole string) error {
	const op = "storage.postgres.SaveUser"

	query := s.sqlBuilder.Insert("users").Columns("full_name", "passhash", "phone", "email", "user_role").Values(fullName, password, phone, email, userRole)
	_, err := query.Exec()

	if err != nil {
		var pqError *pq.Error

		if errors.As(err, &pqError) && pqError.Code == UniqueViolationCode {
			return fmt.Errorf("%s: %w", op, storage.ErrUserExist)
		}

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) UserByEmail(email string) (*models.User, error) {
	const op = "storage.postgres.UserByEmail"

	query := s.sqlBuilder.Select("id", "passhash", "user_role", "deleted").From("users").Where(sq.Eq{"email": email})
	rows, err := query.Query()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(rows)

	var user *models.User

	for rows.Next() {
		var (
			id       int64
			passHash string
			userRole string
			deleted  bool
		)
		if err := rows.Scan(&id, &passHash, &userRole, &deleted); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		user = &models.User{Id: id, PassHash: []byte(passHash), RoleString: userRole, Email: email, Deleted: deleted}
	}

	if user == nil || user.Deleted {
		return nil, storage.ErrUserNotFound
	}

	return user, nil
}

func (s *Storage) UserByPhone(phone string) (*models.User, error) {
	const op = "storage.postgres.UserByPhone"

	query := s.sqlBuilder.Select("id", "passhash", "user_role", "email", "deleted").From("users").Where(sq.Eq{"phone": phone})
	rows, err := query.Query()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(rows)

	var user *models.User

	for rows.Next() {
		var (
			id       int64
			passHash string
			userRole string
			email    string
			deleted  bool
		)
		if err := rows.Scan(&id, &passHash, &userRole, &email, &deleted); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		user = &models.User{Id: id, PassHash: []byte(passHash), RoleString: userRole, Email: email, Deleted: deleted}
	}

	if user == nil || user.Deleted {
		return nil, storage.ErrUserNotFound
	}

	return user, nil
}

func (s *Storage) ForgetPasswordInfo(link string) (*models.ForgetPasswordInfo, error) {
	const op = "storage.postgres.ForgetPasswordInfo"

	query := s.sqlBuilder.Select("id", "link", "user_id", "expiration").From("forget_password_info").Where(sq.Eq{"link": link})
	rows, err := query.Query()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(rows)

	var forgetPasswordInfo *models.ForgetPasswordInfo

	for rows.Next() {
		var (
			id         int64
			link       string
			userId     int64
			expiration time.Time
		)
		if err := rows.Scan(&id, &link, &userId, &expiration); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		forgetPasswordInfo = &models.ForgetPasswordInfo{Id: id, Link: link, UserId: userId, Expiration: expiration}
	}

	if forgetPasswordInfo == nil {
		return nil, storage.LinkNotFound
	}

	return forgetPasswordInfo, nil
}

func (s *Storage) DeleteLinkById(id int64) error {
	const op = "storage.postgres.DeleteLinkById"

	query := s.sqlBuilder.Delete("forget_password_info").Where(sq.Eq{"id": id})

	_, err := query.Exec()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) UpdatePassword(userId int64, newPassword string) error {
	const op = "storage.postgres.UpdatePassword"

	query := s.sqlBuilder.Update("users").Set("passhash", newPassword).Where(sq.Eq{"id": userId})
	_, err := query.Exec()

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) UserByUserId(userId int64) (*models.User, error) {
	const op = "storage.postgres.UserByEmail"

	query := s.sqlBuilder.Select("id", "passhash", "user_role", "email", "deleted").From("users").Where(sq.Eq{"id": userId})
	rows, err := query.Query()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(rows)

	var user *models.User

	for rows.Next() {
		var (
			id       int64
			passHash string
			userRole string
			email    string
			deleted  bool
		)
		if err := rows.Scan(&id, &passHash, &userRole, &email, &deleted); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		user = &models.User{Id: id, PassHash: []byte(passHash), RoleString: userRole}
	}

	if user == nil || user.Deleted {
		return nil, storage.ErrUserNotFound
	}

	return user, nil
}

func (s *Storage) SaveLink(link string, linkTtl time.Duration, userId int64) error {
	const op = "storage.postgres.SaveUser"

	query := s.sqlBuilder.Insert("forget_password_info").Columns("link", "user_id", "expiration").Values(link, userId, time.Now().Add(linkTtl))
	_, err := query.Exec()

	if err != nil {
		var pqError *pq.Error

		if errors.As(err, &pqError) && pqError.Code == UniqueViolationCode {
			return fmt.Errorf("%s: %w", op, storage.ErrUserExist)
		}

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) UpdateUser(newFullName, newPhone, newEmail string, userId int64) error {
	const op = "storage.postgres.UpdateUser"

	query := s.sqlBuilder.Update("users").Set("full_name", newFullName).Set("phone", newPhone).Set("email", newEmail).Where(sq.Eq{"id": userId})
	_, err := query.Exec()

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) DeleteUser(userId int64) error {
	const op = "storage.postgres.DeleteUser"

	query := s.sqlBuilder.Update("users").Set("deleted", true).Where(sq.Eq{"id": userId})
	_, err := query.Exec()

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) RestoreUser(userId int64) error {
	const op = "storage.postgres.DeleteUser"

	query := s.sqlBuilder.Update("users").Set("deleted", false).Where(sq.Eq{"id": userId})
	_, err := query.Exec()

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

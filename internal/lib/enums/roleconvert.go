package enums

import "auth/internal/domain/models"

const (
	roleAdmin     = "admin"
	roleJobSeeker = "jobseeker"
	roleEmployer  = "employer"
)

func RoleConvertFromString(role string) models.UserRole {
	if role == roleAdmin {
		return models.Admin
	} else if role == roleJobSeeker {
		return models.JobSeeker
	} else if role == roleEmployer {
		return models.Employer
	} else {
		return 0
	}
}

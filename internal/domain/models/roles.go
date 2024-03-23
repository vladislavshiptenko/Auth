package models

type UserRole int

const (
	JobSeeker UserRole = iota + 1
	Employer
	Admin
)

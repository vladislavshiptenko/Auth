package models

import "time"

type ForgetPasswordInfo struct {
	Id         int64
	Link       string
	UserId     int64
	Expiration time.Time
}

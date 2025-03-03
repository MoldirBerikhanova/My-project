package models

import "time"

type User struct {
	Id              int        `form:"id"`
	Name            string     `form:"name"`
	Email           string     `form:"email"`
	Password        string     `form:"password"`
	ConfirmPassword *string    `form:"confirm_password"`
	PhoneNumber     *string    `form:"phonenumber"`
	Birthday        *time.Time `form:"birthday"`
	IsAdmin         *string    `form:"is_admin"`
	PosterUrl       *string    `form:"posterUrl"`
}

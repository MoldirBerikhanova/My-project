package models

import "time"

type User struct {
	Id          int        `form:"id"`
	Name        string     `form:"name"`
	Email       string     `form:"email"`
	Password    string     `form:"password"`
	PhoneNumber *string    `form:"phonenumber"`
	Birthday    *time.Time `form:"birthday"`
	IsAdmin     *string    `form:"role"`
	PosterUrl   *string    `form:"posterUrl"`
}

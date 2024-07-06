package models

import (
	"fmt"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	FirstName     string
	LastName      string
	Email         string `gorm:"unique"`
	Password      string
	Phone         string
	Organisations []*Organisation `gorm:"many2many:user_organistaions;"`
}

type UserRegisterParams struct {
	FirstName string `json:"firstName" validate:"required,alpha,min=2,max=32"`
	LastName  string `json:"lastName" validate:"required,alpha,min=2,max=32"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8,max=32"`
	Phone     string `json:"phone" validate:"required,len=11"`
}

type UserLoginParams struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=32"`
}

func UserResponse(user User) map[string]string {
	return map[string]string{
		"userId":    fmt.Sprintf("%d", user.ID),
		"firstName": user.FirstName,
		"lastName":  user.LastName,
		"email":     user.Email,
		"phone":     user.Phone,
	}
}

package models

import (
	"fmt"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	FirstName            string
	LastName             string
	Email                string `gorm:"unique"`
	Password             string `json:"-"`
	Phone                string
	CreatedOrganisations []Organisation `gorm:"foreignKey:CreatedByID"`
	Organisations        []Organisation `gorm:"many2many:users_organisations"`
}

type UserRegisterParams struct {
	FirstName string `json:"firstName" validate:"required,min=1,max=64"`
	LastName  string `json:"lastName" validate:"required,min=1,max=64"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=1,max=64"`
	Phone     string `json:"phone" validate:"required,min=1"`
}

type UserLoginParams struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=1,max=64"`
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

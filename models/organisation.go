package models

import "gorm.io/gorm"

type Organisation struct {
	gorm.Model
	Name   string
	UserID int
	User   User
	Users  []*User `gorm:"many2many:user_organistaions;"`
}

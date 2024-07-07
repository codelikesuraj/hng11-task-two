package models

import (
	"fmt"

	"gorm.io/gorm"
)

type Organisation struct {
	gorm.Model
	Name        string
	Description string
	CreatedByID uint
	CreatedBy   User   `gorm:"foreignKey:CreatedByID"`
	Users       []User `gorm:"many2many:users_organisations"`
}

func OrganisationsResponse(organisations []Organisation) []map[string]string {
	orgs := []map[string]string{}

	if len(organisations) < 1 {
		return orgs
	}

	for _, org := range organisations {
		orgs = append(orgs, map[string]string{
			"orgId":       fmt.Sprintf("%d", org.ID),
			"name":        org.Name,
			"description": org.Description,
		})
	}
	
	return orgs
}

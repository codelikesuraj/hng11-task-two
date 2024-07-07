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

type OrganisationCreateParams struct {
	Name        string `json:"name" validate:"required,min=1,max=64"`
	Description string `json:"description" validate:"omitempty,min=1,max=64"`
}

type OrganisationUserParams struct {
	UserID string `json:"userId" validate:"required"`
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

func OrganisationResponse(organisation Organisation) map[string]string {
	return map[string]string{
		"orgId":       fmt.Sprintf("%d", organisation.ID),
		"name":        organisation.Name,
		"description": organisation.Description,
	}
}

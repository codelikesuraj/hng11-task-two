package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/codelikesuraj/hng11-task-two/models"
	"github.com/codelikesuraj/hng11-task-two/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

type OrganisationController struct {
	DB *gorm.DB
}

func NewOrganisationController(db *gorm.DB) *OrganisationController {
	return &OrganisationController{DB: db}
}

func (oc *OrganisationController) Create(c *gin.Context) {
	var org models.OrganisationCreateParams
	validate := validator.New(validator.WithRequiredStructEnabled())

	if err := c.ShouldBind(&org); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     http.StatusText(http.StatusInternalServerError),
			"message":    err.Error(),
			"statusCode": http.StatusInternalServerError,
		})
		return
	}

	if err := validate.Struct(org); err != nil {
		ve := err.(validator.ValidationErrors)
		errors := make([]models.InputError, len(ve))
		for i, fe := range ve {
			errors[i] = models.InputError{
				Field:   utils.GetJSONTagValue(org, fe.Field()),
				Message: utils.GetValidationMessage(fe),
			}
		}
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"errors": errors,
		})
		return
	}

	tokenString, _ := utils.GetJWTFromRequest(c)
	userFromJWT, err := utils.GetUserFromJWT(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":     http.StatusText(http.StatusUnauthorized),
			"message":    http.StatusText(http.StatusUnauthorized),
			"statusCode": http.StatusUnauthorized,
		})
		return
	}

	userJWTId, _ := strconv.ParseUint(userFromJWT["userId"].(string), 10, 64)
	var user models.User

	if result := oc.DB.Limit(1).First(&user, userJWTId); result.Error != nil || result.RowsAffected < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     http.StatusText(http.StatusBadRequest),
			"message":    "invalid user",
			"statusCode": http.StatusBadRequest,
		})
		return
	}

	newOrg := models.Organisation{
		Name:        org.Name,
		Description: org.Description,
		CreatedByID: user.ID,
	}

	err = oc.DB.Transaction(func(tx *gorm.DB) error {
		// create organisation
		if err := tx.Create(&newOrg).Error; err != nil {
			return err
		}

		// add user to organisation
		if err := tx.Model(&newOrg).Association("Users").Append(&user); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     http.StatusText(http.StatusInternalServerError),
			"message":    http.StatusText(http.StatusInternalServerError),
			"statusCode": http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  http.StatusText(http.StatusCreated),
		"message": "Organisation created successfully",
		"data":    models.OrganisationResponse(newOrg),
	})
}

func (oc *OrganisationController) GetAll(c *gin.Context) {
	tokenString, _ := utils.GetJWTFromRequest(c)
	userFromJWT, err := utils.GetUserFromJWT(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":     http.StatusText(http.StatusUnauthorized),
			"message":    http.StatusText(http.StatusUnauthorized),
			"statusCode": http.StatusUnauthorized,
		})
		return
	}

	// userJWTId, _ := strconv.ParseUint(userJWT["userId"].(string), 10, 64)
	// if uint(userJWTId) != user.ID {
	// 	c.JSON(http.StatusUnauthorized, gin.H{
	// 		"status":     http.StatusText(http.StatusUnauthorized),
	// 		"message":    http.StatusText(http.StatusUnauthorized),
	// 		"statusCode": http.StatusUnauthorized,
	// 	})
	// 	return
	// }

	var user models.User

	result := oc.DB.Where("id = ?", userFromJWT["userId"]).Preload("Organisations").Limit(1).First(&user)
	if result.RowsAffected < 1 || result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":     http.StatusText(http.StatusNotFound),
			"message":    "user not found",
			"statusCode": http.StatusNotFound,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": fmt.Sprintf("found %d organisation(s)", len(user.Organisations)),
		"data": gin.H{
			"organisations": models.OrganisationsResponse(user.Organisations),
		},
	})
}

func (oc *OrganisationController) GetOrganisationById(c *gin.Context) {
	orgId, _ := strconv.Atoi(c.Param("orgId"))
	if orgId < 1 {
		c.JSON(http.StatusNotFound, gin.H{
			"status":     http.StatusText(http.StatusNotFound),
			"message":    "invalid organisation ID",
			"statusCode": http.StatusNotFound,
		})
		return
	}

	tokenString, _ := utils.GetJWTFromRequest(c)
	userFromJWT, err := utils.GetUserFromJWT(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":     http.StatusText(http.StatusUnauthorized),
			"message":    http.StatusText(http.StatusUnauthorized),
			"statusCode": http.StatusUnauthorized,
		})
		return
	}

	var user models.User
	var orgs []models.Organisation

	result := oc.DB.Where("id = ?", userFromJWT["userId"]).Preload("Organisations").Limit(1).First(&user)
	if result.RowsAffected < 1 || result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":     http.StatusText(http.StatusNotFound),
			"message":    "user not found",
			"statusCode": http.StatusNotFound,
		})
		return
	}

	err = oc.DB.Model(&user).Where("id = ?", orgId).Association("Organisations").Find(&orgs)
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound) || len(orgs) < 1:
		c.JSON(http.StatusNotFound, gin.H{
			"status":     http.StatusText(http.StatusNotFound),
			"message":    "organisation not found",
			"statusCode": http.StatusNotFound,
		})
		return
	case err != nil:
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     http.StatusText(http.StatusInternalServerError),
			"message":    http.StatusText(http.StatusInternalServerError),
			"statusCode": http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "found organisation",
		"data": gin.H{
			"organisations": models.OrganisationResponse(orgs[0]),
		},
	})
}

func (oc *OrganisationController) AddUser(c *gin.Context) {
	// validate userID parameter
	var addUser models.OrganisationUserParams
	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := c.ShouldBind(&addUser); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     http.StatusText(http.StatusInternalServerError),
			"message":    err.Error(),
			"statusCode": http.StatusInternalServerError,
		})
		return
	}

	if err := validate.Struct(addUser); err != nil {
		ve := err.(validator.ValidationErrors)
		errors := make([]models.InputError, len(ve))
		for i, fe := range ve {
			errors[i] = models.InputError{
				Field:   utils.GetJSONTagValue(addUser, fe.Field()),
				Message: utils.GetValidationMessage(fe),
			}
		}
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"errors": errors,
		})
		return
	}

	// check if userId exists
	var newUser models.User
	result := oc.DB.Limit(1).Find(&newUser, addUser.UserID)
	if result.RowsAffected < 1 || result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":     http.StatusText(http.StatusNotFound),
			"message":    "user not found",
			"statusCode": http.StatusNotFound,
		})
		return
	}

	// get orgId
	orgId, _ := strconv.Atoi(c.Param("orgId"))
	if orgId < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     http.StatusText(http.StatusBadRequest),
			"message":    "invalid organisation ID",
			"statusCode": http.StatusBadRequest,
		})
		return
	}
	var org models.Organisation
	result = oc.DB.Limit(1).Find(&org, orgId)
	if result.RowsAffected < 1 || result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":     http.StatusText(http.StatusNotFound),
			"message":    "organisation not found",
			"statusCode": http.StatusNotFound,
		})
		return
	}

	// get authUser
	tokenString, _ := utils.GetJWTFromRequest(c)
	userFromJWT, err := utils.GetUserFromJWT(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":     http.StatusText(http.StatusUnauthorized),
			"message":    http.StatusText(http.StatusUnauthorized),
			"statusCode": http.StatusUnauthorized,
		})
		return
	}

	var authUser models.User

	result = oc.DB.Limit(1).Find(&authUser, userFromJWT["userId"])
	if result.RowsAffected < 1 || result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":     http.StatusText(http.StatusNotFound),
			"message":    "user not found",
			"statusCode": http.StatusNotFound,
		})
		return
	}

	// check if auth user has access to organisation
	if oc.DB.Model(&authUser).Where("id = ?", org.ID).Association("Organisations").Count() < 1 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":     http.StatusText(http.StatusUnauthorized),
			"message":    "user cannot access organisation",
			"statusCode": http.StatusUnauthorized,
		})
		return
	}

	// add user to org
	if err := oc.DB.Model(&org).Association("Users").Append(&newUser); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     http.StatusText(http.StatusInternalServerError),
			"message":    "error adding user to organisation",
			"statusCode": http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "User added to organisation successfully",
	})
}

package controllers

import (
	"log"
	"net/http"

	"github.com/codelikesuraj/hng11-task-two/models"
	"github.com/codelikesuraj/hng11-task-two/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

type UserController struct {
	DB *gorm.DB
}

func NewUserController(db *gorm.DB) *UserController {
	return &UserController{DB: db}
}

func (uc *UserController) RegisterUser(c *gin.Context) {
	var user models.UserRegisterParams
	validate := validator.New(validator.WithRequiredStructEnabled())

	if err := c.ShouldBind(&user); err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     http.StatusText(http.StatusInternalServerError),
			"message":    err.Error(),
			"statusCode": http.StatusInternalServerError,
		})
		return
	}

	if err := validate.Struct(user); err != nil {
		ve := err.(validator.ValidationErrors)
		errors := make([]models.InputError, len(ve))
		for i, fe := range ve {
			log.Println(fe)
			errors[i] = models.InputError{
				Field:   utils.GetJSONTagValue(user, fe.Field()),
				Message: utils.GetValidationMessage(fe),
			}
		}
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"errors": errors,
		})
		return
	}

	passwordHash, err := utils.HashPassword(user.Password)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     http.StatusText(http.StatusInternalServerError),
			"message":    err.Error(),
			"statusCode": http.StatusInternalServerError,
		})
		return
	}

	newUser := models.User{
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		Password:  passwordHash,
		Phone:     user.Phone,
	}

	if err = uc.DB.Create(&newUser).Error; err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     http.StatusText(http.StatusInternalServerError),
			"message":    err.Error(),
			"statusCode": http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "registration successful",
		"data": gin.H{
			"accessToken": "<access_token>",
			"user":        models.UserResponse(newUser),
		},
	})
}

func (uc *UserController) LoginUser(c *gin.Context) {
	var userParam models.UserLoginParams
	validate := validator.New(validator.WithRequiredStructEnabled())

	if err := c.ShouldBind(&userParam); err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     http.StatusText(http.StatusInternalServerError),
			"message":    err.Error(),
			"statusCode": http.StatusInternalServerError,
		})
		return
	}

	if err := validate.Struct(userParam); err != nil {
		ve := err.(validator.ValidationErrors)
		errors := make([]models.InputError, len(ve))
		for i, fe := range ve {
			log.Println(fe)
			errors[i] = models.InputError{
				Field:   utils.GetJSONTagValue(userParam, fe.Field()),
				Message: utils.GetValidationMessage(fe),
			}
		}
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"errors": errors,
		})
		return
	}

	var user models.User
	result := uc.DB.Where("email = ?", userParam.Email).Limit(1).Find(&user)
	if err := result.Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     http.StatusText(http.StatusInternalServerError),
			"message":    err.Error(),
			"statusCode": http.StatusInternalServerError,
		})
		return
	}

	if result.RowsAffected < 1 || !utils.PasswordIsValid(user.Password, userParam.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":     "bad request",
			"message":    "authentication failed",
			"statusCode": http.StatusUnauthorized,
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "login successful",
		"data": gin.H{
			"accessToken": "<access_token>",
			"users":       models.UserResponse(user),
		},
	})
}

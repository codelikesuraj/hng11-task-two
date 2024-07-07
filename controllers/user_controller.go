package controllers

import (
	"fmt"
	"net/http"
	"strconv"

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
	result := uc.DB.Where("email = ?", newUser.Email).Limit(1).Find(&newUser)
	if err := result.Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     http.StatusText(http.StatusInternalServerError),
			"message":    err.Error(),
			"statusCode": http.StatusInternalServerError,
		})
		return
	}
	if result.RowsAffected > 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"errors": []gin.H{
				{
					"field":   "email",
					"message": "email already exists",
				},
			},
		})
		return
	}

	newUser, err = registerUserWithOrg(uc.DB, newUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     http.StatusText(http.StatusInternalServerError),
			"message":    err.Error(),
			"statusCode": http.StatusInternalServerError,
		})
		return
	}

	token, err := utils.GenerateJWT(newUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     http.StatusText(http.StatusInternalServerError),
			"message":    err.Error(),
			"statusCode": http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "Registration successful",
		"data": gin.H{
			"accessToken": token,
			"user":        models.UserResponse(newUser),
		},
	})
}

func (uc *UserController) LoginUser(c *gin.Context) {
	var userParam models.UserLoginParams
	validate := validator.New(validator.WithRequiredStructEnabled())

	if err := c.ShouldBind(&userParam); err != nil {
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
			"status":     "Bad request",
			"message":    "Authentication failed",
			"statusCode": http.StatusUnauthorized,
		})
		return
	}

	token, err := utils.GenerateJWT(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     http.StatusText(http.StatusInternalServerError),
			"message":    err.Error(),
			"statusCode": http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Login successful",
		"data": gin.H{
			"accessToken": token,
			"user":        models.UserResponse(user),
		},
	})
}

func (uc *UserController) GetUserById(c *gin.Context) {
	userId, _ := strconv.Atoi(c.Param("id"))
	if userId < 1 {
		c.JSON(http.StatusNotFound, gin.H{
			"status":     http.StatusText(http.StatusNotFound),
			"message":    "user not found",
			"statusCode": http.StatusNotFound,
		})
		return
	}

	var user models.User
	user.ID = uint(userId)

	tokenString, _ := utils.GetJWTFromRequest(c)
	_, err := utils.GetUserFromJWT(tokenString)
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

	result := uc.DB.Limit(1).Find(&user)
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
		"message": "user found",
		"data":    models.UserResponse(user),
	})
}

func registerUserWithOrg(db *gorm.DB, user models.User) (models.User, error) {
	err := db.Transaction(func(tx *gorm.DB) error {
		// create user
		if err := tx.Create(&user).Error; err != nil {
			return err
		}

		// create organisation
		org := models.Organisation{
			Name:        fmt.Sprintf("%s's Organisation", user.FirstName),
			CreatedByID: user.ID,
		}
		if err := tx.Create(&org).Error; err != nil {
			return err
		}

		// add user to organisation
		if err := tx.Model(&org).Association("Users").Append(&user); err != nil {
			return err
		}

		return nil
	})

	return user, err
}

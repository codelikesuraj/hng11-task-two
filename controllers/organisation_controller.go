package controllers

import (
	"fmt"
	"net/http"

	"github.com/codelikesuraj/hng11-task-two/models"
	"github.com/codelikesuraj/hng11-task-two/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type OrganisationController struct {
	DB *gorm.DB
}

func NewOrganisationController(db *gorm.DB) *OrganisationController {
	return &OrganisationController{DB: db}
}

func (uc *OrganisationController) GetAll(c *gin.Context) {
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

	result := uc.DB.Where("id = ?", userFromJWT["userId"]).Preload("Organisations").Limit(1).First(&user)
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

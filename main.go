package main

import (
	"log"
	"os"

	"github.com/codelikesuraj/hng11-task-two/controllers"
	"github.com/codelikesuraj/hng11-task-two/middlewares"
	"github.com/codelikesuraj/hng11-task-two/models"
	"github.com/codelikesuraj/hng11-task-two/utils"
	"github.com/gin-gonic/gin"
)

func init() {
	utils.LoadEnvs()
}

func main() {
	// initialize database
	db, err := utils.GetDBConnection()
	if err != nil {
		log.Fatal("error connecting to database:", err)
	}
	db.AutoMigrate(models.User{}, &models.Organisation{})

	router := gin.Default()

	UserController := controllers.NewUserController(db)

	router.GET("/", controllers.Home)
	router.GET("/auth", middlewares.Auth(), controllers.Home)
	router.POST("/auth/register", UserController.RegisterUser)
	router.POST("/auth/login", UserController.LoginUser)
	router.GET("/api/users/:id", middlewares.Auth(), UserController.GetUserById)

	router.Run(":" + os.Getenv("PORT"))
}

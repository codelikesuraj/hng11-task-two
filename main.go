package main

import (
	"log"
	"os"

	"github.com/codelikesuraj/hng11-task-two/controllers"
	"github.com/codelikesuraj/hng11-task-two/middlewares"
	"github.com/codelikesuraj/hng11-task-two/models"
	"github.com/codelikesuraj/hng11-task-two/utils"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func init() {
	godotenv.Load()
}

func main() {
	// initialize database
	db, err := utils.GetDBConnection()
	if err != nil {
		log.Fatal("error connecting to database:", err)
	}
	db.AutoMigrate(&models.Organisation{}, models.User{})

	OrganisationController := controllers.NewOrganisationController(db)
	UserController := controllers.NewUserController(db)

	router := gin.Default()
	router.GET("/", controllers.Home)
	router.Group("/auth").
		POST("/register", UserController.RegisterUser).
		POST("/login", UserController.LoginUser)
	router.Group("/api", middlewares.Auth()).
		GET("/users/:id", UserController.GetUserById).
		GET("/organisations/:orgId", OrganisationController.GetOrganisationById).
		GET("/organisations", OrganisationController.GetAll).
		POST("/organisations", OrganisationController.Create).
		POST("/organisations/:orgId/users", OrganisationController.AddUser)
	router.Run(":" + os.Getenv("PORT"))
}

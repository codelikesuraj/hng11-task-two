package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/codelikesuraj/hng11-task-two/controllers"
	"github.com/codelikesuraj/hng11-task-two/middlewares"
	"github.com/codelikesuraj/hng11-task-two/models"
	"github.com/codelikesuraj/hng11-task-two/utils"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

const (
	letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	numberBytes = "0123456789"
)

var db *gorm.DB

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func GenerateRandomNumber() string {
	b := make([]byte, 11)
	for i := range b {
		b[i] = numberBytes[rand.Intn(len(numberBytes))]
	}
	return string(b)
}

func GenerateRandomString(n int) string {
	return fmt.Sprint(RandStringBytes(n))
}

func GenerateRandomEmail() string {
	return fmt.Sprintf("%s@example.com", RandStringBytes(10))
}

func init() {
	godotenv.Load()

	var err error
	db, err = utils.GetDBConnection()
	if err != nil {
		log.Fatal("error connecting to database:", err)
	}
	db.AutoMigrate(&models.Organisation{}, models.User{})
}

func setupRouter() *gin.Engine {
	userController := controllers.UserController{DB: db}
	organisationController := controllers.OrganisationController{DB: db}

	router := gin.New()
	router.GET("/", controllers.Home)
	authRoutes := router.Group("/auth")
	{
		authRoutes.POST("/register", userController.RegisterUser)
		authRoutes.POST("/login", userController.LoginUser)
	}
	apiRoutes := router.Group("/api", middlewares.Auth())
	{
		apiRoutes.GET("/users/:id", userController.GetUserById)
		apiRoutes.GET("/organisations/:orgId", organisationController.GetOrganisationById)
		apiRoutes.GET("/organisations", organisationController.GetAll)
		apiRoutes.POST("/organisations", organisationController.Create)
		apiRoutes.POST("/organisations/:orgId/users", organisationController.AddUser)
	}
	return router
}

func TestHomeRoute(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"message":"this is the default page"}`, w.Body.String())
}

func TestAuthRoutes(t *testing.T) {
	router := setupRouter()

	t.Run("test user cannot register on validation error", func(t *testing.T) {
		registerParamsJSON, _ := json.Marshal(map[string]string{})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(registerParamsJSON))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnprocessableEntity, w.Code, w.Body.String())
	})

	t.Run("test user can register successfully", func(t *testing.T) {
		registerParamsJSON, _ := json.Marshal(map[string]string{
			"firstName": GenerateRandomString(10),
			"lastName":  GenerateRandomString(10),
			"email":     GenerateRandomEmail(),
			"phone":     GenerateRandomNumber(),
			"password":  GenerateRandomString(8),
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(registerParamsJSON))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code, w.Body.String())
	})

	t.Run("test user cannot login on validation error", func(t *testing.T) {
		loginParamsJSON, _ := json.Marshal(map[string]string{})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(loginParamsJSON))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnprocessableEntity, w.Code, w.Body.String())
	})

	t.Run("test user cannot login with invalid details", func(t *testing.T) {
		loginParamsJSON, _ := json.Marshal(map[string]string{
			"email":    GenerateRandomEmail(),
			"password": GenerateRandomString(8),
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(loginParamsJSON))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code, w.Body.String())
	})

	t.Run("test user can login successfully", func(t *testing.T) {
		firstName := GenerateRandomString(5)
		lastName := GenerateRandomString(5)
		email := GenerateRandomEmail()
		phone := GenerateRandomNumber()
		password := GenerateRandomString(8)

		registerParamsJSON, _ := json.Marshal(map[string]string{
			"firstName": firstName,
			"lastName":  lastName,
			"email":     email,
			"phone":     phone,
			"password":  password,
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(registerParamsJSON))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code, w.Body.String())

		loginParamsJSON, _ := json.Marshal(map[string]string{
			"email":    email,
			"password": password,
		})
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("POST", "/auth/login", bytes.NewBuffer(loginParamsJSON))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code, w.Body.String())
	})
}

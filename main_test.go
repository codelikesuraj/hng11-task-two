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
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type RegisterSuccessResponse struct {
	Data struct {
		AccessToken string `json:"accessToken"`
		User        struct {
			Email     string `json:"email"`
			FirstName string `json:"firstName"`
			LastName  string `json:"lastName"`
			Phone     string `json:"phone"`
			UserID    string `json:"userId"`
		} `json:"user"`
	} `json:"data"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

const (
	letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	numberBytes = "0123456789"
)

var (
	db     *gorm.DB
	router *gin.Engine
)

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

	router = setupRouter()
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
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"message":"this is the default page"}`, w.Body.String())
}

func TestAuthRoutes(t *testing.T) {
	t.Run("test user cannot register on validation error", func(t *testing.T) {
		registerParamsJSON, _ := json.Marshal(map[string]string{})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(registerParamsJSON))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnprocessableEntity, w.Code, w.Body.String())
	})

	t.Run("test user cannot register with existing email", func(t *testing.T) {
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
		require.Equal(t, http.StatusCreated, w.Code, w.Body.String())

		w = httptest.NewRecorder()
		req, _ = http.NewRequest("POST", "/auth/register", bytes.NewBuffer(registerParamsJSON))
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
		require.Equal(t, http.StatusCreated, w.Code, w.Body.String())

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

func TestAPIRoutes(t *testing.T) {
	var resp RegisterSuccessResponse

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
	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())

	err := json.NewDecoder(w.Body).Decode(&resp)
	require.Nil(t, err, err)

	testToken := resp.Data.AccessToken
	// testUser := resp.Data.User

	t.Run("test unauthenticated user cannot access routes", func(t *testing.T) {
		randNum := GenerateRandomNumber()
		testCases := []struct {
			method, url string
		}{
			{"GET", "/api/users/" + randNum},
			{"GET", "/api/organisations/" + randNum},
			{"GET", "/api/organisations"},
			{"POST", "/api/organisations"},
			{"POST", "/api/organisations/" + randNum + "/users"},
		}

		for _, testCase := range testCases {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(testCase.method, testCase.url, nil)
			router.ServeHTTP(w, req)
			assert.NotEqual(t, http.StatusInternalServerError, w.Code, resp)
			assert.Equal(t, http.StatusUnauthorized, w.Code, fmt.Sprintln(w.Body.String(), ":", testCase.url))
		}
	})

	t.Run("test authenticate user can access routes", func(t *testing.T) {
		emptyParams, _ := json.Marshal(map[string]string{})
		emptyParamsJson := bytes.NewBuffer(emptyParams)

		randNum := GenerateRandomNumber()

		testCases := []struct {
			method, url string
			body        *bytes.Buffer
		}{
			{"GET", "/api/users/" + randNum, emptyParamsJson},
			{"GET", "/api/organisations/" + randNum, emptyParamsJson},
			{"GET", "/api/organisations", emptyParamsJson},
			{"POST", "/api/organisations", emptyParamsJson},
			// {"POST", "/api/organisations/" + "1" + "/users", emptyParamsJson},
		}

		for _, testCase := range testCases {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(testCase.method, testCase.url, emptyParamsJson)
			req.Header.Set("Authorization", "Bearer "+testToken)
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)
			resp := fmt.Sprintln(w.Body.String(), ":", testCase.url)
			assert.NotEqual(t, http.StatusInternalServerError, w.Code, resp)
			assert.Equal(t, http.StatusUnauthorized, w.Code, resp)
		}
	})
}

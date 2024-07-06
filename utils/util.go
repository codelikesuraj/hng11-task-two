package utils

import (
	"errors"
	"log"
	"os"
	"reflect"
	"time"

	"github.com/codelikesuraj/hng11-task-two/models"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v4"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func GetDBConnection() (*gorm.DB, error) {
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN: os.Getenv("PG_URL"),
	}))
	if err != nil {
		return nil, err
	}

	return db, nil
}

func GetValidationMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "alpha":
		return "field must be only contain alphabets"
	case "email":
		return "invalid email"
	case "required":
		return "field is required"
	case "min":
		return "must be at least " + fe.Param() + " characters long"
	case "max":
		return "must not be more than " + fe.Param() + " characters"
	case "len":
		return "field must be exactly " + fe.Param() + " characters"
	default:
		return fe.Error()
	}
}

func HashPassword(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", nil
	}
	return string(b), nil
}

func PasswordIsValid(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func GetJSONTagValue(v interface{}, fieldName string) string {
	t := reflect.TypeOf(v)

	if t.Kind() != reflect.Struct {
		return ""
	}

	field, ok := t.FieldByName(fieldName)
	if !ok {
		return ""
	}

	return field.Tag.Get("json")
}

func LoadEnvs() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func GenerateJWT(user models.User) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":   user.ID,
		"user": models.UserResponse(user),
		"exp":  time.Now().Add(time.Hour * 24).Unix(),
	}).SignedString([]byte(os.Getenv("JWT_SECRET")))
}

func CheckJWT(tokenString string) error {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// check signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid siging method")
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	claims, ok := token.Claims.(jwt.MapClaims)

	// check token validity
	if !ok || !token.Valid || err != nil {
		return errors.New("invalid token")
	}

	// check expiry
	if float64(time.Now().Unix()) > claims["exp"].(float64) {
		return errors.New("expired token")
	}

	return nil
}

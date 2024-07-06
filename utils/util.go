package utils

import (
	"log"
	"os"
	"reflect"

	"github.com/go-playground/validator/v10"
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

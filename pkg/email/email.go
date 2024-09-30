package email

import (
	"errors"
	"fmt"
	"regexp"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

var (
	ErrInvalidEmailAddress = errors.New("invalid email address")
)

type Email struct{ Value string }

func isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

func NewEmail(s string) (Email, error) {
	if !isValidEmail(s) {
		return Email{}, ErrInvalidEmailAddress
	}
	return Email{s}, nil
}

func (e Email) MarshalBSONValue() (bsontype.Type, []byte, error) {
	// Marshal the Email as a simple string value in BSON
	return bson.TypeString, bsoncore.AppendString(nil, e.Value), nil
}

func (e *Email) UnmarshalBSONValue(t bsontype.Type, data []byte) error {
	// Check if the type is a string
	if t != bson.TypeString {
		return fmt.Errorf("unexpected BSON type for email: %s", t.String())
	}

	// Unmarshal the data as a string
	e.Value = bsoncore.Value{Type: t, Data: data}.StringValue()
	return nil
}

package email

import (
	"errors"
	"regexp"

	"go.mongodb.org/mongo-driver/bson"
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

func (e Email) MarshalBSON() ([]byte, error) {
	return bson.Marshal(bson.M{"email": e.Value})
}

func (e *Email) UnmarshalBSON(data []byte) error {
	var doc struct {
		Email string `bson:"email"`
	}
	if err := bson.Unmarshal(data, &doc); err != nil {
		return err
	}
	*e = Email{doc.Email}
	return nil
}

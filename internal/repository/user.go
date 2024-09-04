package repository

import (
	"context"
	"errors"
	"fmt"

	. "github.com/We-ll-think-about-it-later/identity-service/internal/model"
	"github.com/We-ll-think-about-it-later/identity-service/pkg/logger"
	"github.com/We-ll-think-about-it-later/identity-service/pkg/mongodb"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	ErrCantFindUserByEmail    = errors.New("can't find user by email")
	ErrUserNotFound           = errors.New("user not found")
	ErrFailedToFindUser       = errors.New("failed to find user")
	ErrEmailAlreadyRegistered = errors.New("email already registered")
	ErrFailedToAddUser        = errors.New("failed to add user")
)

type UserRepository interface {
	CreateUser(ctx context.Context, profileInfo ProfileInfo) (User, error)
	FindByEmail(ctx context.Context, email string) (User, error)
	FindByID(ctx context.Context, userId uuid.UUID) (User, error)
	ConfirmUser(ctx context.Context, userId uuid.UUID) error
}

type UserRepositoryImpl struct {
	*mongodb.Client
	*mongo.Collection
	logger *logger.Logger
}

func NewUserRepository(
	m *mongodb.Client,
	dbName,
	collectionName string,
	logger *logger.Logger,
) UserRepositoryImpl {
	collection := m.Database(dbName).Collection(collectionName)
	logger.SetPrefix("user repository ")

	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "profile_info.email", Value: 1}}, // Index on "email" field in ascending order
		Options: options.Index().SetUnique(true),               // Make the index unique
	}

	// Create a unique index on the "email" field
	indexName, err := collection.Indexes().CreateOne(m.Ctx, indexModel)
	if err != nil {
		// Check if the error is due to a duplicate index
		if mongo.IsDuplicateKeyError(err) {
			logger.Debug("Index \"email\" already exists")
		} else {
			logger.Fatal(err)
		}
	} else {
		logger.Debug("Created index:", indexName)
	}

	return UserRepositoryImpl{m, collection, logger}
}

func (r UserRepositoryImpl) CreateUser(ctx context.Context, profileInfo ProfileInfo) (User, error) {
	user := NewUser(profileInfo)
	insertResult, err := r.InsertOne(ctx, user)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return User{}, ErrEmailAlreadyRegistered
		} else {
			return User{}, ErrFailedToAddUser
		}
	}
	r.logger.Debug("Inserted user: ", insertResult.InsertedID)
	return user, nil
}

func (r UserRepositoryImpl) FindByEmail(ctx context.Context, email string) (User, error) {
	var user User
	err := r.FindOne(
		ctx,
		bson.M{"profile_info.email": email},
	).Decode(&user)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return User{}, ErrUserNotFound
		}
		r.logger.Debugf("failed to find user by email: %v", err)
		return User{}, ErrFailedToFindUser
	}
	return user, nil
}

func (r UserRepositoryImpl) FindByID(ctx context.Context, userId uuid.UUID) (User, error) {
	var user User
	err := r.FindOne(
		ctx,
		bson.M{"_id": userId},
	).Decode(&user)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return User{}, ErrUserNotFound
		}
		return User{}, ErrFailedToAddUser
	}
	return user, nil
}

func (r UserRepositoryImpl) ConfirmUser(ctx context.Context, userId uuid.UUID) error {
	filter := bson.M{"_id": userId}
	update := bson.M{"$set": bson.M{"is_confirmed": true}}

	_, err := r.UpdateOne(ctx, filter, update)
	fmt.Println(err)
	// if err != nil {
	// 	return fmt.Errorf("failed to update user confirmation status: %w", err)
	// }

	return nil
}

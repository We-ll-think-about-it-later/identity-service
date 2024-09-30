package repository

import (
	"errors"
	"fmt"

	"github.com/We-ll-think-about-it-later/identity-service/internal/model"
	"github.com/We-ll-think-about-it-later/identity-service/pkg/mongodb"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	ErrCantUpsertRefreshToken = errors.New("can't upsert refresh token")
	ErrCantFindRefreshToken   = errors.New("can't find refresh token")
)

type TokenRepository interface {
	UpsertRefreshToken(uuid.UUID, model.Encrypted[model.RefreshToken]) error
	FindRefreshToken(uuid.UUID) (model.Encrypted[model.RefreshToken], error)
}

type TokenRepositoryImpl struct {
	*mongodb.Client
	*mongo.Collection
	logger *logrus.Logger
}

func NewTokenRepository(
	m *mongodb.Client,
	dbName,
	collectionName string,
	logger *logrus.Logger,
) TokenRepositoryImpl {
	collection := m.Database(dbName).Collection(collectionName)
	logger = logger.WithField("prefix", "token repository").Logger

	// Create index on user_id
	indexModel := mongo.IndexModel{
		Keys:    bson.M{"user_id": 1},            // Ascending index on user_id
		Options: options.Index().SetUnique(true), // Make the index unique
	}
	_, err := collection.Indexes().CreateOne(m.Ctx, indexModel)
	if err != nil {
		// Check if the error is due to a duplicate index
		if mongo.IsDuplicateKeyError(err) {
			logger.Debug("Index \"user_id\" already exists")
		} else {
			logger.Fatal(err)
		}
	}

	return TokenRepositoryImpl{m, collection, logger}
}

func (r TokenRepositoryImpl) UpsertRefreshToken(userId uuid.UUID, refreshToken model.Encrypted[model.RefreshToken]) error {
	filter := bson.M{"user_id": userId}
	update := bson.M{"$set": bson.M{"refresh_token": refreshToken.String()}}
	upsert := true

	res, err := r.Collection.UpdateOne(r.Ctx, filter, update, &options.UpdateOptions{Upsert: &upsert})
	if err != nil {
		r.logger.Debug(err)
		return fmt.Errorf("failed to upsert refresh token: %w", err)
	}

	if res.ModifiedCount != 1 && res.UpsertedCount != 1 {
		return ErrCantUpsertRefreshToken
	}

	return nil
}

func (r TokenRepositoryImpl) FindRefreshToken(userId uuid.UUID) (model.Encrypted[model.RefreshToken], error) {
	var result struct {
		UserID       string `bson:"user_id"`
		RefreshToken string `bson:"refresh_token"`
	}

	filter := bson.M{"user_id": userId}
	err := r.Collection.FindOne(r.Ctx, filter).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return model.Encrypted[model.RefreshToken]{}, ErrCantFindRefreshToken
		}
		r.logger.Debug(err)
		return model.Encrypted[model.RefreshToken]{}, fmt.Errorf("failed to find refresh token: %w", err)
	}

	encrypted, err := model.EncryptedFromString[model.RefreshToken](result.RefreshToken)
	if err != nil {
		r.logger.Debug(err)
		return model.Encrypted[model.RefreshToken]{}, fmt.Errorf("failed to decode refresh token: %w", err)
	}

	return encrypted, nil
}

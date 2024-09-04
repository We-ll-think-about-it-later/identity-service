package repository

import (
	"context"
	"errors"

	. "github.com/We-ll-think-about-it-later/identity-service/internal/model"
	"github.com/We-ll-think-about-it-later/identity-service/pkg/logger"
	"github.com/We-ll-think-about-it-later/identity-service/pkg/mongodb"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var ErrCodeNotFound = errors.New("code not found")

type CodeRepository interface {
	SaveConfirmationCode(ctx context.Context, userId uuid.UUID, code Encrypted[Code]) error
	GetConfirmationCode(ctx context.Context, userId uuid.UUID) (Encrypted[Code], error)
	DeleteConfirmationCode(ctx context.Context, userId uuid.UUID) error
}

type CodeRepositoryImpl struct {
	*mongodb.Client
	*mongo.Collection
	logger *logger.Logger
}

func NewCodeRepository(
	m *mongodb.Client,
	dbName,
	collectionName string,
	logger *logger.Logger,
) CodeRepositoryImpl {
	collection := m.Database(dbName).Collection(collectionName)
	logger.SetPrefix("confirmation code repository ")

	indexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "user_id", Value: 1},
		},
		Options: options.Index().SetExpireAfterSeconds(300), // Expire after 5 minutes
	}
	_, err := collection.Indexes().CreateOne(m.Ctx, indexModel)
	if err != nil {
		logger.Fatalf("failed to create TTL index: %v", err)
	}

	return CodeRepositoryImpl{m, collection, logger}
}

func (r CodeRepositoryImpl) SaveConfirmationCode(ctx context.Context, userId uuid.UUID, code Encrypted[Code]) error {
	filter := bson.M{"user_id": userId}
	update := bson.M{
		"$set": bson.M{
			"user_id": userId,
			"code":    code.String(),
		},
	}
	opts := options.Update().SetUpsert(true)

	_, err := r.Collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		r.logger.Errorf("failed to save confirmation code: %v", err)
		return err
	}

	return nil
}

func (r CodeRepositoryImpl) GetConfirmationCode(ctx context.Context, userId uuid.UUID) (Encrypted[Code], error) {
	var result struct {
		HashedCode string `bson:"code"`
	}

	err := r.Collection.FindOne(
		ctx,
		bson.M{"user_id": userId},
	).Decode(&result)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return Encrypted[Code]{}, ErrCodeNotFound
		}
		r.logger.Debugf("failed to get confirmation code: %v", err)
		return Encrypted[Code]{}, err
	}

	code, _ := EncryptedFromString[Code](result.HashedCode)
	return code, nil
}

func (r CodeRepositoryImpl) DeleteConfirmationCode(ctx context.Context, userId uuid.UUID) error {
	_, err := r.Collection.DeleteOne(
		ctx,
		bson.M{"user_id": userId},
	)
	return err
}

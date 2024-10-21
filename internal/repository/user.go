package repository

import (
	"context"
	"errors"

	. "github.com/We-ll-think-about-it-later/identity-service/internal/model"
	"github.com/We-ll-think-about-it-later/identity-service/pkg/email"
	"github.com/We-ll-think-about-it-later/identity-service/pkg/mongodb"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
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
	CreateUser(ctx context.Context, email email.Email) (User, error)
	FindByEmail(ctx context.Context, email email.Email) (User, error)
	FindById(ctx context.Context, userId uuid.UUID) (User, error)
	CreateProfile(ctx context.Context, userId uuid.UUID, profileInfoUpdate ProfileInfo) error
	UpdateProfile(ctx context.Context, userId uuid.UUID, profileInfoUpdate ProfileInfoUpdate) (ProfileInfo, error)
	FuzzySearchByUsername(ctx context.Context, searchTerm string) ([]UserSearchResult, error)
}

type UserRepositoryImpl struct {
	*mongodb.Client
	*mongo.Collection
	logger *logrus.Logger
}

func NewUserRepository(
	m *mongodb.Client,
	dbName,
	collectionName string,
	logger *logrus.Logger,
) UserRepositoryImpl {
	collection := m.Database(dbName).Collection(collectionName)
	logger = logger.WithField("prefix", "user repository").Logger

	// Create a unique index on the "email" field
	emailIndexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}}, // Index on "email" field in ascending order
		Options: options.Index().SetUnique(true),  // Make the index unique
	}

	emailIndexName, err := collection.Indexes().CreateOne(m.Ctx, emailIndexModel)
	if err != nil {
		// Check if the error is due to a duplicate index
		if mongo.IsDuplicateKeyError(err) {
			logger.Debug("Index \"email\" already exists")
		} else {
			logger.Fatal(err)
		}
	} else {
		logger.Debug("Created index:", emailIndexName)
	}

	// Create Atlas Search index
	searchIndexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "profile_info.username", Value: "text"},
		},
		Options: options.Index().SetName("profileUsernameIndex"),
	}

	searchIndexName, err := collection.Indexes().CreateOne(m.Ctx, searchIndexModel)
	if err != nil {
		// Check if the error is due to a duplicate index
		if mongo.IsDuplicateKeyError(err) {
			logger.Debug("Atlas Search index \"profileUsernameIndex\" already exists")
		} else {
			logger.Fatal(err)
		}
	} else {
		logger.Debug("Created Atlas Search index:", searchIndexName)
	}

	return UserRepositoryImpl{m, collection, logger}
}

func (r UserRepositoryImpl) CreateUser(ctx context.Context, email email.Email) (User, error) {
	user := NewUser(email)
	_, err := r.InsertOne(ctx, user)
	return user, err
}

func (r UserRepositoryImpl) FindByEmail(ctx context.Context, email email.Email) (User, error) {
	var user User
	err := r.FindOne(
		ctx,
		bson.M{"email": email.Value},
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

func (r UserRepositoryImpl) FindById(ctx context.Context, userId uuid.UUID) (User, error) {
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

func (r UserRepositoryImpl) CreateProfile(ctx context.Context, userId uuid.UUID, profileInfo ProfileInfo) error {
	filter := bson.M{"_id": userId}
	update := bson.M{"$set": bson.M{"profile_info": profileInfo}}

	_, err := r.UpdateOne(ctx, filter, update)
	// if errors.Is(err, ErrNothingFiltered){
	// 	return ErrUserNotFound
	// }
	if err != nil {
		return err
	}
	return nil
}

func (r UserRepositoryImpl) UpdateProfile(ctx context.Context, userId uuid.UUID, profileInfoUpdate ProfileInfoUpdate) (ProfileInfo, error) {
	filter := bson.M{"_id": userId}

	update := bson.M{"$set": bson.M{}}
	if profileInfoUpdate.FirstName != nil {
		update["$set"].(bson.M)["profile_info.firstname"] = *profileInfoUpdate.FirstName
	}
	if profileInfoUpdate.LastName != nil {
		update["$set"].(bson.M)["profile_info.lastname"] = *profileInfoUpdate.LastName
	}
	if profileInfoUpdate.UserName != nil {
		update["$set"].(bson.M)["profile_info.username"] = *profileInfoUpdate.UserName
	}

	_, err := r.UpdateOne(ctx, filter, update)
	// if errors.Is(err, ErrNothingFiltered){
	// 	return ErrUserNotFound
	// }
	if err != nil {
		return ProfileInfo{}, err
	}

	var user User
	err = r.FindOne(ctx, bson.M{"_id": userId}).Decode(&user)
	if err != nil {
		return ProfileInfo{}, err
	}
	return *user.ProfileInfo, err
}

func (r UserRepositoryImpl) FuzzySearchByUsername(ctx context.Context, searchTerm string) ([]UserSearchResult, error) {

	// Define the aggregation pipeline
	pipeline := []bson.M{
		{
			"$search": bson.M{
				"index": "profileUsernameIndex", // Atlas Search index name
				"text": bson.M{
					"query": searchTerm,
					"path":  "profile_info.username",
					"fuzzy": bson.M{}, // Enable fuzzy matching
				},
			},
		},
		{
			"$project": bson.M{
				"_id":          1,
				"profile_info": 1,
				"score":        bson.M{"$meta": "searchScore"}, // Include the search score
			},
		},
		{
			"$sort": bson.M{"score": -1}, // Sort by score in descending order
		},
	}

	// Execute the aggregation pipeline
	cursor, err := r.Aggregate(ctx, pipeline)
	if err != nil {
		r.logger.Errorf("failed to execute aggregation pipeline: %v", err)
		return nil, err
	}

	// Decode the results
	var results []UserSearchResult
	if err := cursor.All(ctx, &results); err != nil {
		r.logger.Errorf("failed to decode aggregation results: %v", err)
		return nil, err
	}

	return results, nil
}

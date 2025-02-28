package mongo

import (
	"context"

	"github.com/FM1337/flagbrew-website-source-code/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const colUsers = "users"

// UserService satisfies the models.UserService interface.
type userSrv struct {
	srv *mongoSrv
}

func (s *mongoSrv) NewUserService() *userSrv {
	return &userSrv{srv: s}
}

func (s *userSrv) Upsert(ctx context.Context, user *models.User) (err error) {

	// if the user ID isn't ""
	if !user.ID.IsZero() {
		_, err = s.srv.users.UpdateOne(ctx, bson.M{"_id": user.ID}, bson.M{"$set": user}, options.Update().SetUpsert(true))
		return errorWrapper(err)
	}

	// if the ID isn't set do we need to set an ID?
	//user.ID = primitive.NewObjectID()

	result, err := s.srv.users.UpdateOne(ctx, bson.M{"github_id": user.GithubID}, bson.M{"$set": user}, options.Update().SetUpsert(true))
	if result.UpsertedID == nil {
		// get the ID from the existing record
		tmpUser := &models.User{}
		if err = s.srv.users.FindOne(ctx, bson.M{"github_id": user.GithubID}).Decode(&tmpUser); err != nil {
			return errorWrapper(err)
		}
		user.ID = tmpUser.ID
	} else {
		user.ID = result.UpsertedID.(primitive.ObjectID)
	}
	return errorWrapper(err)
}

func (s *userSrv) Get(ctx context.Context, id string) (user *models.User, err error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errorWrapper(err)
	}

	result := s.srv.users.FindOne(ctx, bson.M{"_id": oid})
	if result.Err() != nil {
		return user, errorWrapper(result.Err())
	}
	err = result.Decode(&user)
	return user, errorWrapper(err)
}

func (s *userSrv) GitHubExists(ctx context.Context, id int) (err error) {

	exists, err := s.srv.users.CountDocuments(ctx, bson.M{"github_id": id})
	if err != nil {
		return errorWrapper(err)
	}
	if exists == 0 {
		return errorWrapper(mongo.ErrNoDocuments)
	}
	return errorWrapper(err)
}

func (s *userSrv) Exists(ctx context.Context, id string) (err error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errorWrapper(err)
	}

	exists, err := s.srv.users.CountDocuments(ctx, bson.M{"_id": oid})
	if err != nil {
		return errorWrapper(err)
	}
	if exists == 0 {
		return errorWrapper(mongo.ErrNoDocuments)
	}
	return errorWrapper(err)
}

func (s *userSrv) Delete(ctx context.Context, id string) (err error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errorWrapper(err)
	}

	_, err = s.srv.users.DeleteOne(ctx, bson.M{"_id": oid})
	return errorWrapper(err)
}

func (s *userSrv) List(ctx context.Context) (users []*models.User, err error) {
	results, err := s.srv.users.Find(ctx, bson.M{})
	if err != nil {
		return users, errorWrapper(err)
	}
	err = results.All(ctx, &users)
	return users, errorWrapper(err)
}

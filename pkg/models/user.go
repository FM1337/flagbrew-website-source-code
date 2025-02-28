package models

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserService interface {
	Upsert(ctx context.Context, r *User) error
	Get(ctx context.Context, id string) (*User, error)
	Exists(ctx context.Context, id string) error
	GitHubExists(ctx context.Context, id int) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]*User, error)
}

type User struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	GithubID int                `bson:"github_id" validate:"required" json:"github_id"`
	Token    string             `bson:"token" json:"-"`

	AvatarURL string `bson:"avatar_url" json:"avatar_url"`
	Username  string `bson:"username" json:"username"`
	Name      string `bson:"name" json:"name"`
	Email     string `bson:"email" validate:"email" json:"-"`

	AccountCreated time.Time `bson:"account_created" json:"-"`
	AccountUpdated time.Time `bson:"account_updated" json:"-"`
}

func (r *User) Validate() error {
	return validateStruct(r)
}

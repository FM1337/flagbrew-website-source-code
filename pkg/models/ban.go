package models

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

type BanService interface {
	Ban(ctx context.Context, ban *Ban) error
	Unban(ctx context.Context, ip string) error
	ListBans(ctx context.Context, query bson.M, page, limit int, sort bson.M) ([]*Ban, int, int64, error)
	ListBan(ctx context.Context, query bson.M) (*Ban, error)
}

type Ban struct {
	Date      time.Time `bson:"date" json:"date"`
	IP        string    `bson:"ip" json:"ip"`
	BanReason string    `bson:"ban_reason" json:"ban_reason"`
	BannedBy  string    `bson:"banned_by" json:"banned_by"`
}

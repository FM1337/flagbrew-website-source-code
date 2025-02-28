package models

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
)

type ApprovalService interface {
	Approve(ctx context.Context, code, user string) (err error)
	ApproveBundle(ctx context.Context, code, user string) (err error)
	Reject(ctx context.Context, code, reason, user string) (err error)
	ListPending(ctx context.Context, query bson.M, page, limit int, sort bson.M) (pending []*GPSSPokemon, pages int, count int64, err error)
}

type RestrictService interface {
	RestrictUploader(ctx context.Context, ip, reason, user string) (err error)
	UnrestrictUploader(ctx context.Context, ip, user string) (err error)
	ListRestricted(ctx context.Context, query bson.M, page, limit int, sort bson.M) (restricted []*RestrictedUploader, pages int, count int64, err error)
	IsUploaderRestricted(ctx context.Context, ip string) (bool, string)
}

type RestrictedUploader struct {
	IP               string `bson:"ip" json:"ip"`
	RestrictedReason string `bson:"restricted_reason" json:"restricted_reason"`
	RestrictedBy     string `bson:"restricted_by" json:"restricted_by"`
}

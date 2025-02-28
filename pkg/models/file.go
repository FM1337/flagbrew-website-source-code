package models

import (
	"bytes"
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FileService interface {
	UploadMysteryGift(ctx context.Context, filename string, data []byte) (size int, err error)
	DownloadMysteryGift(ctx context.Context, filename string) (size int64, file bytes.Buffer, err error)

	UploadPatreonBuild(ctx context.Context, filename string, data []byte, hash, app, extension string) (size int, err error)
	DownloadPatreonBuild(ctx context.Context, app, hash, extension string) (size int64, file bytes.Buffer, err error)
	PatreonBuildExists(ctx context.Context, app, hash, extension string) bool
	GetLatestAppHash(ctx context.Context, app string) (hash interface{}, err error)
	CleanOldBuilds(ctx context.Context, app, hash string) (deleted []*File, err error)
}

type File struct {
	ID         primitive.ObjectID `bson:"_id"`
	Length     int64              `bson:"length"`
	ChunkSize  int32              `bson:"chunkSize"`
	UploadDate time.Time          `bson:"uploadDate"`
	Filename   string             `bson:"filename"`
	Metadata   Metadata           `bson:"metadata"`
}

type Metadata struct {
	CommitHash string    `bson:"commit_hash"`
	App        string    `bson:"app"`
	Extension  string    `bson:"extension"`
	ExpireDate time.Time `bson:"expire_date"`
}

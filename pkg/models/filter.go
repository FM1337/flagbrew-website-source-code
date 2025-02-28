package models

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

type FilterService interface {
	ListWords(ctx context.Context, query bson.M, page, limit int, sort bson.M) ([]*WordFilter, int, int64, error)
	ListLegality(ctx context.Context, query bson.M, page, limit int, sort bson.M) ([]*LegalityFilter, int, int64, error)
	AddWord(ctx context.Context, word string, strict, caseInsensitive bool, createdBy string) (err error)
	RemoveWord(ctx context.Context, word string, caseInsensitive bool) (err error)
	CheckWords(ctx context.Context, words []string, mode string, caseInsensitive bool) (match, strict bool, index int, err error)
}

type WordFilter struct {
	AddedBy       string    `bson:"added_by" json:"added_by"`
	CreatedDate   time.Time `bson:"created_date" json:"created_date"`
	String        string    `bson:"string" json:"string"` // The string to look for
	Strict        bool      `bson:"strict" json:"strict"` // If strict, Pokemon will be auto rejected rather than being held for review
	CaseSensitive bool      `bson:"case_sensitive" json:"case_sensitive"`
}

type LegalityFilter struct {
	String string `bson:"string" json:"string"` // The string to look for
}

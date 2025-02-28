package models

import "context"

type Patron struct {
	Code      string `bson:"code" json:"code"`
	DiscordID string `bson:"discord_id" json:"discord_id"`
}

type PatronService interface {
	IsPatron(ctx context.Context, code string) bool
	GetPatronDiscord(ctx context.Context, code string) (id string, err error)
}

type PatreonCleanupDaemon interface {
	Start()
	Stop()
}

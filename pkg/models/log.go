package models

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

type LogService interface {
	UpsertLog(ctx context.Context, l *interface{}) (err error)
	ListLogs(ctx context.Context, query bson.M, page, limit int, sort bson.M) (logs interface{}, pages int, count int64, err error)
	GetLog(ctx context.Context, query bson.M) (log interface{}, err error)
	UpdateLog(ctx context.Context, query, update bson.M) (err error)
	DeleteLog(ctx context.Context, query bson.M) (err error)
}

// ValidLogTypes contains a list of valid log types for the get log route
var ValidLogTypes = map[string]bool{"gpss_upload": true, "gpss_failed_upload": true, "gpss_deletion": true, "discord_warn": true, "unban": true, "banned": true,
	"setting_change": true, "gpss_clean": true, "build_delete": true, "unrestrict": true, "restrictions": true, "gpss_bundle_upload": true, "word_delete": true}

type GPSSUploadLog struct {
	Date            time.Time `bson:"date" json:"date"`
	UploaderIP      string    `bson:"uploader_ip" json:"uploader_ip"`
	UploadSource    string    `bson:"upload_source" json:"upload_source"`
	UploaderDiscord string    `bson:"uploader_discord,omitempty" json:"uploader_discord,omitempty"` // Check header for discord
	Deleted         bool      `bson:"deleted" json:"deleted"`
	PokemonData     Pokemon   `bson:"pokemon_data" json:"pokemon_data"`
	Approved        bool      `bson:"approved" json:"approved"`                                   // True by default, false by default when approval system implemented and enabled
	ApprovedBy      string    `bson:"approved_by" json:"approved_by"`                             // System by default, moderator name when approval system implemented and enabled, N/A when not approved
	Rejected        bool      `bson:"rejected" json:"rejected"`                                   // false by default, requires approval system
	RejectedBy      string    `bson:"rejected_by,omitempty" json:"rejected_by,omitempty"`         // blank by default, requires approval system
	RejectedReason  string    `bson:"rejected_reason,omitempty" json:"rejected_reason,omitempty"` // blank by default, requires approval system
	Patron          bool      `bson:"patron" json:"patron"`
	PatronCode      string    `bson:"patron_code,omitempty" json:"patron_code,omitempty"`
	PatronDiscord   string    `bson:"patron_discord,omitempty" json:"patron_discord,omitempty"`
	BundleUpload    bool      `bson:"bundle_upload" json:"bundle_upload"`
	DownloadCode    string    `bson:"download_code" json:"download_code"`
	BundleCode      string    `bson:"bundle_code,omitempty" json:"bundle_code,omitempty"`
	LogType         string    `bson:"log_type" json:"log_type"` // Defines the log type, should be "gpss_upload" for this one
	DBVersion       int       `bson:"db_version" json:"db_version"`
}

// type GPSSUploadLog struct {
// 	ID              primitive.ObjectID `bson:"_id" json:"id"`
// 	Date            time.Time          `bson:"date" json:"date"`
// 	UploaderIP      string             `bson:"uploader_ip" json:"uploader_ip"`
// 	UploadSource    string             `bson:"upload_source" json:"upload_source"`
// 	UploaderDiscord string             `bson:"uploader_discord,omitempty" json:"uploader_discord,omitempty"` // Check header for discord
// 	Deleted         bool               `bson:"deleted" json:"deleted"`
// 	PokemonData     Pokemon            `bson:"pokemon_data" json:"pokemon_data"`
// 	Approved        bool               `bson:"approved" json:"approved"`                                   // True by default, false by default when approval system implemented and enabled
// 	ApprovedBy      string             `bson:"approved_by" json:"approved_by"`                             // System by default, moderator name when approval system implemented and enabled, N/A when not approved
// 	Rejected        bool               `bson:"rejected" json:"rejected"`                                   // false by default, requires approval system
// 	RejectedBy      string             `bson:"rejected_by,omitempty" json:"rejected_by,omitempty"`         // blank by default, requires approval system
// 	RejectedReason  string             `bson:"rejected_reason,omitempty" json:"rejected_reason,omitempty"` // blank by default, requires approval system
// 	Patron          bool               `bson:"patron" json:"patron"`
// 	PatronCode      string             `bson:"patron_code,omitempty" json:"patron_code,omitempty"`
// 	PatronDiscord   string             `bson:"patron_discord,omitempty" json:"patron_discord,omitempty"`
// 	BundleUpload    bool               `bson:"bundle_upload" json:"bundle_upload"`
// 	DownloadCode    string             `bson:"download_code" json:"download_code"`
// 	BundleCode      string             `bson:"bundle_code,omitempty" json:"bundle_code,omitempty"`
// 	LogType         string             `bson:"log_type" json:"log_type"` // Defines the log type, should be "gpss_upload" for this one
// }

type GPSSBundleUploadLog struct {
	Date            time.Time `bson:"date" json:"date"`
	UploaderIP      string    `bson:"uploader_ip" json:"uploader_ip"`
	UploadSource    string    `bson:"upload_source" json:"upload_source"`
	UploaderDiscord string    `bson:"uploader_discord,omitempty" json:"uploader_discord,omitempty"` // Check header for discord
	Deleted         bool      `bson:"deleted" json:"deleted"`
	Pokemons        []Pokemon `bson:"pokemons" json:"pokemons"`
	Patron          bool      `bson:"patron" json:"patron"`
	PatronCode      string    `bson:"patron_code,omitempty" json:"patron_code,omitempty"`
	PatronDiscord   string    `bson:"patron_discord,omitempty" json:"patron_discord,omitempty"`
	DownloadCodes   []string  `bson:"download_codes" json:"download_codes"`
	DownloadCode    string    `bson:"download_code" json:"download_code"`
	Approved        bool      `bson:"approved" json:"approved"`
	LogType         string    `bson:"log_type" json:"log_type"` // Defines the log type, should be "gpss_bundle_upload" for this one
	DBVersion       int       `bson:"db_version" json:"db_version"`
}

// type GPSSBundleUploadLog struct {
// 	ID              primitive.ObjectID `bson:"_id" json:"id"`
// 	Date            time.Time          `bson:"date" json:"date"`
// 	UploaderIP      string             `bson:"uploader_ip" json:"uploader_ip"`
// 	UploadSource    string             `bson:"upload_source" json:"upload_source"`
// 	UploaderDiscord string             `bson:"uploader_discord,omitempty" json:"uploader_discord,omitempty"` // Check header for discord
// 	Deleted         bool               `bson:"deleted" json:"deleted"`
// 	Pokemons        []Pokemon          `bson:"pokemons" json:"pokemons"`
// 	Patron          bool               `bson:"patron" json:"patron"`
// 	PatronCode      string             `bson:"patron_code,omitempty" json:"patron_code,omitempty"`
// 	PatronDiscord   string             `bson:"patron_discord,omitempty" json:"patron_discord,omitempty"`
// 	DownloadCodes   []string           `bson:"download_codes" json:"download_codes"`
// 	DownloadCode    string             `bson:"download_code" json:"download_code"`
// 	Approved        bool               `bson:"approved" json:"approved"`
// 	LogType         string             `bson:"log_type" json:"log_type"` // Defines the log type, should be "gpss_bundle_upload" for this one
// }

type GPSSFailedUploadLog struct {
	Date            time.Time `bson:"date" json:"date"`
	UploaderIP      string    `bson:"uploader_ip" json:"uploader_ip"`
	UploadSource    string    `bson:"upload_source" json:"upload_source"`
	UploaderDiscord string    `bson:"uploader_discord,omitempty" json:"uploader_discord,omitempty"` // Check header for discord
	FailedReason    string    `bson:"rejected_reason" json:"rejected_reason"`
	Patron          bool      `bson:"patron" json:"patron"`
	PatronCode      string    `bson:"patron_code,omitempty" json:"patron_code,omitempty"`
	PatronDiscord   string    `bson:"patron_discord,omitempty" json:"patron_discord,omitempty"`
	LogType         string    `bson:"log_type" json:"log_type"` // Defines the log type, should be "gpss_failed_upload" for this one
}

type GPSSDeletionLog struct {
	Date           time.Time `bson:"date" json:"date"`
	DeletedBy      string    `bson:"deleted_by" json:"deleted_by"`
	DeletionReason string    `bson:"deletion_reason" json:"deletion_reason"`
	EntityType     string    `bson:"entity_type" json:"entity_type"` // Pokemon/Bundle
	DownloadCode   string    `bson:"download_code" json:"download_code"`
	LogType        string    `bson:"log_type" json:"log_type"` // Defines the log type, should be "gpss_deletion" for this one
}

type DiscordWarns struct {
	User    string `json:"user" bson:"user"`
	Warns   []Warn `bson:"warns" json:"warns"`
	LogType string `bson:"log_type" json:"log_type"` // Defines the log type, should be "discord_warn" for this one
}

type Warn struct {
	Date     time.Time `bson:"date" json:"date"`
	Reason   string    `bson:"reason" json:"reason"`
	WarnedBy string    `bson:"warned_by" json:"warned_by"`
}

type UnbanLog struct {
	Date       time.Time `bson:"date" json:"date"`
	Ban        Ban       `bson:"ban" json:"ban"`
	UnbannedBy string    `bson:"unbanned_by" json:"unbanned_by"`
	LogType    string    `bson:"log_type" json:"log_type"` // Defines the log type, should be "unban" for this one
}

type SettingChangeLog struct {
	Date          time.Time   `bson:"date" json:"date"`
	Setting       string      `bson:"setting" json:"setting"`
	OriginalValue interface{} `bson:"original_value" json:"original_value"`
	NewValue      interface{} `bson:"new_value" json:"new_value"`
	ModifiedBy    string      `bson:"modified_by" json:"modified_by"`
	LogType       string      `bson:"log_type" json:"log_type"` // Defines the log type, should be "setting_change" for this one
}

type GPSSCleanLog struct {
	Date    time.Time `bson:"date" json:"date"`
	Deleted int64     `bson:"deleted" json:"deleted"`
	Reset   int64     `bson:"reset" json:"reset"`
	Failed  int64     `bson:"failed" json:"failed"`
	LogType string    `bson:"log_type" json:"log_type"` // Defines the log type, should be "gpss_clean" for this one
}

type PatreonBuildDeleteLog struct {
	Date               time.Time `bson:"date" json:"date"`
	CommitHash         string    `bson:"commit_hash" json:"commit_hash"`
	Filename           string    `bson:"filename" json:"filename"`
	OriginalExpiryDate time.Time `bson:"original_expiry_date" json:"original_expiry_date"`
	LogType            string    `bson:"log_type" json:"log_type"` // Defines the log type, should be "build_delete" for this one
}

type UnrestrictedLog struct {
	Date             time.Time          `bson:"date" json:"date"`
	UnrestrictedBy   string             `bson:"unrestricted_by" json:"unrestricted_by"`
	OriginalRestrict RestrictedUploader `bson:"original_restriction" json:"original_restriction"`
	LogType          string             `bson:"log_type" json:"log_type"` // Defines the log type, should be "unrestrict" for this one
}

type WordDeleteLog struct {
	Date         time.Time `bson:"date" json:"date"`
	DeletedBy    string    `bson:"deleted_by" json:"deleted_by"`
	OriginalWord string    `bson:"original_word" json:"original_word"`
	LogType      string    `bson:"log_type" json:"log_type"` // Defines the log type, should be "word_delete" for this one
}

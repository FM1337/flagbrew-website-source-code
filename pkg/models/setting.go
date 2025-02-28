package models

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

type SettingService interface {
	InsertSetting(ctx context.Context, setting *Setting) error
	UpdateSetting(ctx context.Context, name string, value interface{}) error
	DeleteSetting(ctx context.Context, name string) error // Used for deleting NON-SYSTEM settings
	ListSettings(ctx context.Context, query bson.M, page, limit int, sort bson.M) ([]*Setting, int, int64, error)
	LoadDefaults(ctx context.Context) error // Will load all the default settings, used when fresh db
}

type SettingDaemon interface {
	Start()
	Stop()
	GetOwnerIP() string
}

type Setting struct {
	Name           string      `bson:"name" json:"name"`
	MapKey         string      `bson:"map_key" json:"map_key"` // Key used for Golang Map
	Description    string      `bson:"description" json:"description"`
	Type           string      `bson:"type" json:"type"`
	Value          interface{} `bson:"value" json:"value"`
	CanBeEmpty     bool        `bson:"can_be_empty" json:"can_be_empty"`       // Can the value be empty?
	SystemVariable bool        `bson:"system_variable" json:"system_variable"` // SystemVariables are variables that cannot be deleted
	CreatedBy      string      `bson:"created_by" json:"created_by"`           // CreatedBy defaults to System if it is a System Variable
	CreatedDate    time.Time   `bson:"created_date" json:"created_date"`
	ModifiedDate   time.Time   `bson:"modified_date" json:"modified_date"`
}

// DefaultSettings are the defaults used by LoadDefaults()
var DefaultSettings = []Setting{
	{
		Name:           "GPSS Uploading Enabled",
		MapKey:         "gpss_upload_enabled",
		Description:    "Determines if uploading to GPSS is enabled or not",
		Type:           "bool",
		Value:          true,
		SystemVariable: true,
		CreatedBy:      "System",
		CreatedDate:    time.Now(),
		ModifiedDate:   time.Now(),
	},
	{
		Name:           "GPSS Downloading Enabled",
		MapKey:         "gpss_download_enabled",
		Description:    "Determines if downloading from GPSS is enabled or not",
		Type:           "bool",
		Value:          true,
		SystemVariable: true,
		CreatedBy:      "System",
		CreatedDate:    time.Now(),
		ModifiedDate:   time.Now(),
	},
	{
		Name:           "GPSS Deleting Enabled",
		MapKey:         "gpss_clean_enabled",
		Description:    "Determines if automatic cleaning of GPSS is enabled or not",
		Type:           "bool",
		Value:          false,
		SystemVariable: true,
		CreatedBy:      "System",
		CreatedDate:    time.Now(),
		ModifiedDate:   time.Now(),
	},
	{
		Name:           "GPSS Restricted Uploads",
		MapKey:         "gpss_restrict_enabled",
		Description:    "Determines if uploading to GPSS requires manual approvals or not",
		Type:           "bool",
		Value:          false,
		SystemVariable: true,
		CreatedBy:      "System",
		CreatedDate:    time.Now(),
		ModifiedDate:   time.Now(),
	},
	{
		Name:           "Auto Legality Enabled",
		MapKey:         "pksm_auto_legality_enabled",
		Description:    "Determines if PKSM Auto Legality is enabled or not",
		Type:           "bool",
		Value:          false,
		SystemVariable: true,
		CreatedBy:      "System",
		CreatedDate:    time.Now(),
		ModifiedDate:   time.Now(),
	},
	{
		Name:           "Legality Analysis Enabled",
		MapKey:         "pksm_legality_analysis_enabled",
		Description:    "Determines if PKSM Legality Analysis is enabled or not",
		Type:           "bool",
		Value:          false,
		SystemVariable: true,
		CreatedBy:      "System",
		CreatedDate:    time.Now(),
		ModifiedDate:   time.Now(),
	},
	{
		Name:           "Swear API",
		MapKey:         "gpss_swear_api_enabled",
		Description:    "Determines if Uploaded Pokemon should be screened against the external Swear API",
		Type:           "bool",
		Value:          true,
		SystemVariable: true,
		CreatedBy:      "System",
		CreatedDate:    time.Now(),
		ModifiedDate:   time.Now(),
	},
	{
		Name:           "Emergency Mode",
		MapKey:         "emergency_mode",
		Description:    "If enabled, access is locked down to certain IPs only (Allen's Home IP and the Server IP). Used for emergercies/maintence",
		Type:           "bool",
		Value:          true,
		SystemVariable: true,
		CreatedBy:      "System",
		CreatedDate:    time.Now(),
		ModifiedDate:   time.Now(),
	},
	{
		Name:           "Bundles Support",
		MapKey:         "bundles_support",
		Description:    "If disabled, all routes related to bundles will return 503 unavailable.",
		Type:           "bool",
		Value:          true,
		SystemVariable: true,
		CreatedBy:      "System",
		CreatedDate:    time.Now(),
		ModifiedDate:   time.Now(),
	},
}

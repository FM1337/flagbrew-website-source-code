package models

import (
	"context"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type GPSSService interface {
	DownloadPokemon(ctx context.Context, code string, approvedOnly bool) (*GPSSPokemon, error)
	DownloadBundle(ctx context.Context, code string, approvedOnly bool) (bundledPkmn []*GPSSPokemon, err error)
	UpsertPokemon(ctx context.Context, gp *GPSSPokemon, header *http.Header, patron bool, patronCode, patronDiscord string, bundleUpload bool, bundleCode string) (bool, bool, string, error)
	RemovePokemon(ctx context.Context, downloadCode string, rejecting, soft bool) error
	ListPokemons(ctx context.Context, query bson.M, page, limit int, sort bson.M, pksmMode bool) ([]*GPSSPokemon, int, int64, error)
	ListPokemon(ctx context.Context, code string, approvedOnly bool) (*GPSSPokemon, error)
	UpsertBundle(ctx context.Context, b *GPSSBundlePokemon, patron bool, patronCode, patronDiscord, bundleCode string, header *http.Header) (bool, string, error)
	RemoveBundle(ctx context.Context, downloadCode string) error
	ListBundles(ctx context.Context, query bson.M, page, limit int, sort primitive.M) ([]*GPSSBundlePokemon, int, int64, error)
	ListBundle(ctx context.Context, code string) (*GPSSBundlePokemon, error)
	ResetOldPokemonDownloads(ctx context.Context) (modified int64, err error)
	PokemonExists(ctx context.Context, base64 string) (exists bool, code string)
	PokemonApproved(ctx context.Context, code string) (approved bool)
	GetStats(ctx context.Context, approved bool) (pokemon, bundles int64, err error)
	// Migrate(logData, pokemonData []byte, header *http.Header) (err error)
	// NewMigrate() (err error)
	ListCountForFieldStat(ctx context.Context, field string, downloads bool) (generations map[string]float64, err error)
	RandomPokemon(ctx context.Context, amount int, generations []string) ([]*GPSSRandomPokemon, error)
}

type GPSSAPI interface {
	GetPokemonInfo(pokemonFile []byte, formData map[string]string) (*Pokemon, error)
	GetLegalityInfo(pokemon []byte, formData map[string]string) (legality *LegalityInfo, err error)
	AutoLegalize(pokemon []byte, formData map[string]string) (legalize *AutoLegalize, err error)
}

type GPSSCleanupDaemon interface {
	Start()
	Stop()
}

// ValidGPSSEntities contains a list of valid entity types for the deletion route
var ValidGPSSEntities = map[string]bool{"Pokemon": true, "Bundle": true}

// ValidGenerations contains an entry for each valid generation
var ValidGenerations = map[string]float32{
	"1":    1,
	"2":    2,
	"3":    3,
	"4":    4,
	"5":    5,
	"6":    6,
	"7":    7,
	"LGPE": 7.5,
	"8":    8,
	"BDSP": 8.5,
	"PLA":  8.6,
	"9":    9,
}

type GPSSMigrateLog struct {
	IP           string      `json:"ip"`
	Patreon      bool        `json:"patron"`
	UploadDate   StupidThing `json:"upload_date"`
	DownloadCode string      `json:"download_code"`
}

type StupidThing struct {
	DateTime time.Time `json:"$date"`
}

// type GPSSMigratePokemon struct {
// 	Base64         string   `json:"base_64"`
// 	Generation     string   `json:"generation"`
// 	TotalDownloads int      `json:"total_downloads"`
// 	DownloadCode   string   `json:"code"`
// 	GroupCodes     []string `json:"group_codes"`
// }

// type GPSSMigrateBundle struct {
// 	DownloadCodes []string `json:"download_codes"`
// 	DownloadCode  string   `json:"code"`
// 	MinGen        string   `json:"min_gen"`
// 	MaxGen        string   `json:"max_gen"`
// 	Legal         bool     `json:"legal"`
// 	HasLGPE       bool     `json:"has_lgpe"`
// 	Patreon       bool     `json:"patron"`
// }

type Pokemon struct {
	Ability         string        `bson:"ability" json:"ability"`
	AbilityNum      int64         `bson:"ability_num" json:"ability_num"`
	Ball            string        `bson:"ball" json:"ball"`
	Checksum        int64         `bson:"checksum" json:"checksum"`
	ContestStats    []ContestStat `bson:"contest_stats" json:"contest_stats"`
	DexNumber       int64         `bson:"dex_number" json:"dex_number"`
	Ec              string        `bson:"ec" json:"ec"`
	EggData         EggData       `bson:"egg_data" json:"egg_data"`
	Esv             string        `bson:"esv" json:"esv"`
	Exp             int64         `bson:"exp" json:"exp,omitempty"`
	FatefulFlag     bool          `bson:"fateful_flag" json:"fateful_flag"`
	Form            int64         `bson:"form_num" json:"form_num"`
	Friendship      int64         `bson:"friendship" json:"friendship"`
	Gender          string        `bson:"gender" json:"gender"`
	GenderFlag      int64         `bson:"gender_flag" json:"gender_flag"`
	Generation      int           `bson:"generation" json:"generation"`
	HandlingTrainer string        `bson:"ht" json:"ht"`
	HPType          string        `bson:"hp_type" json:"hp_type"`
	HeldItem        string        `bson:"held_item" json:"held_item"`
	IllegalReasons  string        `bson:"illegal_reasons" json:"illegal_reasons"`
	IsEgg           bool          `bson:"is_egg" json:"is_egg"`
	IsNicknamed     bool          `bson:"is_nicknamed" json:"is_nicknamed"`
	IsShiny         bool          `bson:"is_shiny" json:"is_shiny"`
	ItemNum         int64         `bson:"item_num" json:"item_num"`
	IsLegal         bool          `bson:"is_legal" json:"is_legal"`
	Level           int64         `bson:"level" json:"level"`
	Markings        int64         `bson:"markings" json:"markings"`
	MetData         MetData       `bson:"met_data" json:"met_data"`
	Moves           []PokemonMove `bson:"moves" json:"moves"`
	Nature          string        `bson:"nature" json:"nature"`
	Nickname        string        `bson:"nickname" json:"nickname"`
	Ot              string        `bson:"ot" json:"ot"`
	OTLang          string        `bson:"ot_lang" json:"ot_lang"`
	OTAffection     int64         `bson:"ot_affection" json:"ot_affection"`
	OTGender        string        `bson:"ot_gender" json:"ot_gender"`
	PartySize       int64         `bson:"party_size" json:"party_size"`
	Pid             string        `bson:"pid" json:"pid"`
	PKRSDays        int64         `bson:"pkrs_days" json:"pkrs_days"`
	PKRSStrain      int64         `bson:"pkrs_strain" json:"pkrs_strain"`
	RelearnMoves    []PokemonMove `bson:"relearn_moves" json:"relearn_moves"`
	Ribbons         []string      `bson:"ribbons" json:"ribbons"`
	Sid             int64         `bson:"sid" json:"sid"`
	SpecForm        int64         `bson:"spec_form" json:"spec_form"`
	Species         string        `bson:"species" json:"species"`
	Sprites         Sprites       `bson:"sprites" json:"sprites"`
	Stats           []PokemonStat `bson:"stats" json:"stats"`
	StoredSize      int64         `bson:"stored_size" json:"stored_size"`
	Tid             int64         `bson:"tid" json:"tid"`
	Tsv             int64         `bson:"tsv" json:"tsv"`
	Version         string        `bson:"version" json:"version"`
	VersionNumber   int64         `bson:"version_num" json:"version_num"`
}

// type Pokemon struct {
// 	Ability         string        `bson:"ability" json:"ability"`
// 	AbilityNum      int64         `bson:"ability_num" json:"ability_num"`
// 	Ball            string        `bson:"ball" json:"ball"`
// 	Checksum        int64         `bson:"checksum" json:"checksum"`
// 	ContestStats    []ContestStat `bson:"contest_stats" json:"contest_stats"`
// 	DexNumber       int64         `bson:"dex_number" json:"dex_number"`
// 	Ec              string        `bson:"ec" json:"ec"`
// 	EggData         EggData       `bson:"egg_data" json:"egg_data"`
// 	Esv             string        `bson:"esv" json:"esv"`
// 	Exp             int64         `bson:"exp" json:"exp,omitempty"`
// 	FatefulFlag     bool          `bson:"fateful_flag" json:"fateful_flag"`
// 	Form            int64         `bson:"form_num" json:"form_num"`
// 	Friendship      int64         `bson:"friendship" json:"friendship"`
// 	Gender          string        `bson:"gender" json:"gender,omitempty"`
// 	GenderFlag      int64         `bson:"gender_flag" json:"gender_flag"`
// 	Generation      string        `bson:"generation" json:"generation"`
// 	HandlingTrainer string        `bson:"handling_trainer" json:"ht"`
// 	HPType          string        `bson:"hp_type" json:"hp_type"`
// 	HeldItem        string        `bson:"held_item" json:"held_item"`
// 	IllegalReasons  string        `bson:"illegal_reasons" json:"illegal_reasons"`
// 	IsEgg           bool          `bson:"is_egg" json:"is_egg"`
// 	IsNicknamed     bool          `bson:"is_nicknamed" json:"is_nicknamed"`
// 	IsShiny         bool          `bson:"is_shiny" json:"is_shiny"`
// 	ItemNum         int64         `bson:"item_num" json:"item_num"`
// 	IsLegal         bool          `bson:"is_legal" json:"is_legal"`
// 	Level           int64         `bson:"level" json:"level"`
// 	Markings        int64         `bson:"markings" json:"markings"`
// 	MetData         MetData       `bson:"met_data" json:"met_data"`
// 	Moves           []PokemonMove `bson:"moves" json:"moves"`
// 	Nature          string        `bson:"nature" json:"nature"`
// 	Nickname        string        `bson:"nickname" json:"nickname"`
// 	Ot              string        `bson:"ot" json:"ot"`
// 	OTLang          string        `bson:"ot_lang" json:"ot_lang"`
// 	OTAffection     int64         `bson:"ot_affection" json:"ot_affection"`
// 	OTGender        interface{}   `bson:"ot_gender" json:"ot_gender"`
// 	Pid             string        `bson:"pid" json:"pid"`
// 	PKRSDays        int64         `bson:"pkrs_days" json:"pkrs_days"`
// 	PKRSStrain      int64         `bson:"pkrs_strain" json:"pkrs_strain"`
// 	RelearnMoves    []RelearnMove `bson:"relearn_moves" json:"relearn_moves"`
// 	Ribbons         []string      `bson:"ribbons" json:"ribbons"`
// 	Sid             int64         `bson:"sid" json:"sid"`
// 	Size            int64         `bson:"size" json:"size"`
// 	SpecForm        int64         `bson:"spec_form" json:"spec_form"`
// 	Species         string        `bson:"species" json:"species"`
// 	Sprites         Sprites       `bson:"sprites" json:"sprites"`
// 	Stats           []PokemonStat `bson:"stats" json:"stats"`
// 	Tid             int64         `bson:"tid" json:"tid"`
// 	Tsv             int64         `bson:"tsv" json:"tsv"`
// 	Version         string        `bson:"version" json:"version"`
// }

type ContestStat struct {
	Name  string `bson:"name" json:"name"`
	Value int64  `bson:"value" json:"value"`
}

// type ContestStat struct {
// 	StatName  string `bson:"stat_name" json:"stat_name"`
// 	StatValue int64  `bson:"stat_value" json:"stat_value"`
// }

type EggData struct {
	Day   int64  `bson:"day" json:"day"`
	Month int64  `bson:"month" json:"month"`
	Year  int64  `bson:"year" json:"year"`
	Name  string `bson:"name" json:"name"`
}

// type EggData struct {
// 	Day      int64  `bson:"day" json:"day"`
// 	Month    int64  `bson:"month" json:"month"`
// 	Year     int64  `bson:"year" json:"year"`
// 	Location string `bson:"location" json:"location"`
// }

type MetData struct {
	Day   int64  `bson:"day" json:"day"`
	Month int64  `bson:"month" json:"month"`
	Year  int64  `bson:"year" json:"year"`
	Name  string `bson:"name" json:"name"`
	Level int64  `bson:"level" json:"level"`
}

//	type MetData struct {
//		Day      int64  `bson:"day" json:"day"`
//		Month    int64  `bson:"month" json:"month"`
//		Year     int64  `bson:"year" json:"year"`
//		Location string `bson:"location" json:"location"`
//		Level    int64  `bson:"level" json:"level"`
//	}
type PokemonMove struct {
	Name  string `bson:"name" json:"name"`
	Type  string `bson:"type" json:"type"`
	PP    int    `bson:"pp" json:"pp"`
	PPUps int    `bson:"pp_ups" json:"pp_ups"`
}

// type RelearnMove struct {
// 	Name string `bson:"name" json:"name"`
// 	Type string `bson:"type" json:"type"`
// }

type Sprites struct {
	Item    string `bson:"item" json:"item"`
	Species string `bson:"species" json:"species"`
}

// type PokemonMove struct {
// 	Name string `json:"name" bson:"name"`
// 	Type string `json:"type" bson:"type"`
// 	PP   int    `json:"pp" bson:"pp"`
// 	PPUp int    `json:"pp_ups" bson:"pp_ups"`
// }

type PokemonStat struct {
	Name  string `json:"name" bson:"name"`
	IV    int    `json:"iv" bson:"iv"`
	EV    int    `json:"ev" bson:"ev"`
	Total string `json:"total" bson:"total"`
}

// type PokemonStat struct {
// 	Name  string `json:"stat_name" bson:"stat_name"`
// 	IV    int    `json:"stat_iv" bson:"stat_iv"`
// 	EV    int    `json:"stat_ev" bson:"stat_ev"`
// 	Total string `json:"stat_total" bson:"stat_total"`
// }

// type GPSSPokemon struct {
// 	ID                primitive.ObjectID `bson:"_id,omitempty" json:"id"`
// 	Base64            string             `bson:"base_64" json:"base_64"`
// 	DownloadCode      string             `bson:"download_code" json:"download_code"`
// 	Pokemon           Pokemon            `bson:"pokemon" json:"pokemon"`
// 	Patreon           bool               `bson:"patreon" json:"patreon"`
// 	InGroup           bool               `bson:"in_group" json:"in_group"`
// 	Size              int                `bson:"size" json:"size"`
// 	Generation        string             `bson:"generation" json:"generation"`
// 	LifetimeDownloads int                `bson:"lifetime_downloads" json:"lifetime_downloads"`
// 	CurrentDownloads  int                `bson:"current_downloads" json:"current_downloads"`
// 	DBVersion         int                `bson:"db_version" json:"db_version"`
// 	UploadDate        time.Time          `bson:"upload_date" json:"upload_date"`
// 	LastReset         time.Time          `bson:"last_reset" json:"last_reset"` // Default to upload date
// 	Approved          bool               `bson:"approved" json:"approved"`     // Default to true for now, when approval system in place and enabled set to false by default
// 	Deleted           bool               `bson:"deleted" json:"deleted"`
// }

type GPSSPokemon struct {
	ID                primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Base64            string             `bson:"base_64" json:"base_64"`
	DownloadCode      string             `bson:"download_code" json:"download_code"`
	Pokemon           Pokemon            `bson:"pokemon" json:"pokemon"`
	Patreon           bool               `bson:"patreon" json:"patreon"`
	InGroup           bool               `bson:"in_group" json:"in_group"`
	Size              int                `bson:"size" json:"size"`
	Generation        int                `bson:"generation" json:"generation"`
	LifetimeDownloads int                `bson:"lifetime_downloads" json:"lifetime_downloads"`
	CurrentDownloads  int                `bson:"current_downloads" json:"current_downloads"`
	DBVersion         int                `bson:"db_version" json:"db_version"`
	UploadDate        time.Time          `bson:"upload_date" json:"upload_date"`
	LastReset         time.Time          `bson:"last_reset" json:"last_reset"` // Default to upload date
	Approved          bool               `bson:"approved" json:"approved"`     // Default to true for now, when approval system in place and enabled set to false by default
	Deleted           bool               `bson:"deleted" json:"deleted"`
}

type GPSSBundlePokemon struct {
	ID            primitive.ObjectID      `bson:"_id,omitempty" json:"id"`
	DownloadCount int                     `bson:"download_count" json:"download_count"`
	DownloadCode  string                  `bson:"download_code" json:"download_code"`
	DownloadCodes []string                `bson:"download_codes" json:"download_codes"`
	Pokemons      []GPSSPKSMBundlePokemon `bson:"pokemons" json:"pokemons"`
	UploadDate    time.Time               `bson:"upload_date" json:"upload_date"`
	Patreon       bool                    `bson:"patreon" json:"patreon"`
	MinGen        int                     `bson:"min_gen" json:"min_gen"`
	MaxGen        int                     `bson:"max_gen" json:"max_gen"`
	Count         int                     `bson:"count" json:"count"`
	IsLegal       bool                    `bson:"is_legal" json:"is_legal"`
	Approved      bool                    `bson:"approved" json:"approved"`
	DBVersion     int                     `bson:"db_version" json:"db_version"`
}

// type GPSSBundlePokemon struct {
// 	ID            primitive.ObjectID      `bson:"_id,omitempty" json:"id"`
// 	DownloadCount int                     `bson:"download_count" json:"download_count"`
// 	DownloadCode  string                  `bson:"download_code" json:"download_code"`
// 	DownloadCodes []string                `bson:"download_codes" json:"download_codes"`
// 	Pokemons      []GPSSPKSMBundlePokemon `bson:"pokemons" json:"pokemons"`
// 	UploadDate    time.Time               `bson:"upload_date" json:"upload_date"`
// 	Patreon       bool                    `bson:"patreon" json:"patreon"`
// 	MinGen        string                  `bson:"min_gen" json:"min_gen"`
// 	MaxGen        string                  `bson:"max_gen" json:"max_gen"`
// 	Count         int                     `bson:"count" json:"count"`
// 	IsLegal       bool                    `bson:"is_legal" json:"is_legal"`
// 	Approved      bool                    `bson:"approved" json:"approved"`
// }

type GPSSPKSMBundlePokemon struct {
	Base64     string `bson:"base_64" json:"base_64"`
	Legality   bool   `bson:"legality" json:"legality"`
	Generation int    `bson:"generation" json:"generation"`
}

// type GPSSPKSMBundlePokemon struct {
// 	Base64     string `bson:"base_64" json:"base_64"`
// 	Legality   bool   `bson:"legality" json:"legality"`
// 	Generation string `bson:"generation" json:"generation"`
// }

type LegalityInfo struct {
	Report []string `json:"report"`
	Legal  bool     `json:"legal"`
}

type AutoLegalize struct {
	Pokemon string   `json:"pokemon"`
	Legal   bool     `json:"legal"`
	Report  []string `json:"report"`
	Ran     bool     `json:"ran"`
	Success bool     `json:"success"`
}

type GPSSRandomPokemon struct {
	Base64     string `json:"base_64" bson:"base_64"`
	Generation int    `json:"generation" bson:"generation"`
}

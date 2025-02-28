package helpers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-playground/form"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type gpssPksmPokemonQuery struct {
	MinLevel      int        `form:"min_level,omitempty" json:"min_level,omitempty" validate:"omitempty,min=1,ltefield=MaxLevel"`
	MaxLevel      int        `form:"max_level,omitempty" json:"max_level,omitempty" validate:"omitempty,max=100,gtefield=MinLevel"`
	Generations   []string   `form:"generations,omitempty" json:"generations,omitempty"`
	HoldingItem   bool       `form:"holding_item,omitempty" json:"holding_item,omitempty"`
	Nickname      string     `form:"nickname,omitempty" json:"nickname,omitempty" validate:"max=35"`
	Legal         bool       `form:"legal,omitempty" json:"legal,omitempty"`
	OTName        string     `form:"ot_name,omitempty" json:"ot_name,omitempty" validate:"max=15"`
	OTID          int        `form:"ot_id,omitempty" json:"ot_id,omitempty"`
	HTName        string     `form:"ht_name,omitempty" json:"ht_name,omitempty" validate:"max=15"`
	Operators     []operator `form:"operators,omitempty" json:"operators,omitempty"`
	Species       []string   `form:"species,omitempty" json:"species,omitempty"`
	Mode          string     `form:"mode,omitempty" json:"mode,omitempty" validate:"omitempty,oneof='or' 'and'"`
	SortField     string     `form:"sort_field" json:"sort_field" validate:"required,oneof='latest' 'legality' 'popularity'"`
	SortDirection *bool      `form:"sort_direction" json:"sort_direction" validate:"required"`
	DownloadCode  string     `form:"download_code,omitempty" json:"download_code,omitempty"`
	DownloadCodes string     `form:"download_codes,omitempty" json:"download_codes,omitempty"`
}

type gpssPokemonQuery struct {
	MinLevel      int        `form:"min_level,omitempty" json:"min_level,omitempty" validate:"omitempty,min=1,ltefield=MaxLevel"`
	MaxLevel      int        `form:"max_level,omitempty" json:"max_level,omitempty" validate:"omitempty,max=100,gtefield=MinLevel"`
	Generations   []int      `form:"generations,omitempty" json:"generations,omitempty"`
	HoldingItem   bool       `form:"holding_item,omitempty" json:"holding_item,omitempty"`
	Nickname      string     `form:"nickname,omitempty" json:"nickname,omitempty" validate:"max=35"`
	Legal         bool       `form:"legal,omitempty" json:"legal,omitempty"`
	OTName        string     `form:"ot_name,omitempty" json:"ot_name,omitempty" validate:"max=15"`
	OTID          int        `form:"ot_id,omitempty" json:"ot_id,omitempty"`
	HTName        string     `form:"ht_name,omitempty" json:"ht_name,omitempty" validate:"max=15"`
	Operators     []operator `form:"operators,omitempty" json:"operators,omitempty"`
	Species       []string   `form:"species,omitempty" json:"species,omitempty"`
	Mode          string     `form:"mode,omitempty" json:"mode,omitempty" validate:"omitempty,oneof='or' 'and'"`
	SortField     string     `form:"sort_field" json:"sort_field" validate:"required,oneof='latest' 'legality' 'popularity'"`
	SortDirection *bool      `form:"sort_direction" json:"sort_direction" validate:"required"`
	DownloadCode  string     `form:"download_code,omitempty" json:"download_code,omitempty"`
	DownloadCodes string     `form:"download_codes,omitempty" json:"download_codes,omitempty"`
}

type gpssUploadLogQuery struct {
	LogType string `form:"sort_field" validate:"required,oneof='gpss_upload'"`
}

type operator struct {
	Operator string `json:"operator" validate:"oneof='=' '!=' '>' '<' '>=' '<=' 'IN' 'NOT IN'"`
	Field    string `json:"field" validate:"oneof='generations' 'holding_item' 'nickname' 'ot_name' 'ot_id' 'ht_name' 'species'`
}

// ParseGPSSQuery parses the query we send to the backend from the frontend before sending it to the database for searching
// To make this as compatible as possible, each key we want to be able to search by, should probably match the key name that's in the database
// Mode determines if we search by AND or OR
func ParseGPSSQuery(w http.ResponseWriter, r *http.Request) (query bson.M, sort bson.M, ok bool, generations []int) {

	gpssPQ := gpssPokemonQuery{}
	if r.Header.Get("pksm-mode") == "yes" {
		gpssPKSMPQ := gpssPksmPokemonQuery{}
		if !FDecode(w, r, &gpssPKSMPQ) {
			return bson.M{}, bson.M{}, false, nil // handle error here
		}

		generations := []int{}

		for _, generation := range gpssPKSMPQ.Generations {
			gen, err := strconv.Atoi(generation)
			if err != nil {
				continue
			}

			generations = append(generations, gen)
		}

		gpssPQ = gpssPokemonQuery{
			MinLevel:      gpssPKSMPQ.MinLevel,
			MaxLevel:      gpssPKSMPQ.MaxLevel,
			Generations:   generations,
			HoldingItem:   gpssPKSMPQ.HoldingItem,
			Nickname:      gpssPKSMPQ.Nickname,
			Legal:         gpssPKSMPQ.Legal,
			OTName:        gpssPKSMPQ.OTName,
			OTID:          gpssPKSMPQ.OTID,
			HTName:        gpssPKSMPQ.HTName,
			Operators:     gpssPKSMPQ.Operators,
			Species:       gpssPKSMPQ.Species,
			Mode:          gpssPKSMPQ.Mode,
			SortField:     gpssPKSMPQ.SortField,
			SortDirection: gpssPKSMPQ.SortDirection,
			DownloadCode:  gpssPKSMPQ.DownloadCode,
			DownloadCodes: gpssPKSMPQ.DownloadCodes,
		}
	} else {
		if !FDecode(w, r, &gpssPQ) {
			return bson.M{}, bson.M{}, false, nil // handle error here
		}
	}

	validate := validator.New()

	err := validate.Struct(gpssPQ)
	if err != nil {
		rerr, multiple := validatorErrorFormat(err)
		HttpError(w, r, 400, fmt.Errorf("%s", rerr), multiple, false, false)
		return bson.M{}, bson.M{}, false, nil
	}

	// json.Unmarshal([]byte(gpssPQ.Operators), &gpssOperators)

	for i := range gpssPQ.Operators {
		err = validate.Struct(gpssPQ.Operators[i])
		if err != nil {
			rerr, multiple := validatorErrorFormat(err)
			HttpError(w, r, 400, fmt.Errorf("%s", rerr), multiple, true, true)
			return bson.M{}, bson.M{}, false, nil
		}
	}

	fields := reflect.ValueOf(gpssPQ)
	types := reflect.TypeOf(gpssPQ)
	bsonQueries := []bson.M{}
	for i := 0; i < fields.NumField(); i++ {
		bsonQuery := bson.M{}
		tmpField := strings.Split(types.Field(i).Tag.Get("json"), ",")[0]
		if tmpField == "min_level" || tmpField == "max_level" {
			continue
		} else {
			for _, operator := range gpssPQ.Operators {
				if operator.Field != tmpField {
					continue
				}
				field := ""
				op := ""
				wildcard := false
				switch tmpField {
				case "generations":
					field = "generation"
					break
				case "holding_item":
					field = "pokemon.held_item"
					break
				case "nickname":
					field = "pokemon.nickname"
					break
				case "ot_name":
					field = "pokemon.ot"
					break
				case "ot_id":
					field = "pokemon.tid"
					break
				case "ht_name":
					field = "pokemon.not_ot"
					break
				case "species":
					field = "pokemon.species"
					break
				case "legal":
					field = "pokemon.is_legal"
					break
				case "download_code":
					field = "download_code"
					break
				case "download_codes":
					field = "download_codes"
					break
				default:
					continue
				}

				switch operator.Operator {
				case "=":
					op = "$eq"
					break
				case "!=":
					op = "$ne"
					break
				case ">":
					op = "$gt"
					break
				case ">=":
					op = "$gte"
					break
				case "<":
					op = "$lt"
					break
				case "<=":
					op = "$lte"
					break
				case "IN":
					if fields.Field(i).Type().String() == "string" {
						wildcard = true
						op = ".*" + regexp.QuoteMeta(fields.Field(i).String()) + ".*"
					} else {
						op = "$in"
					}
					break
				case "NOT IN":
					if fields.Field(i).Type().String() == "string" {
						wildcard = true
						op = "^((?!" + regexp.QuoteMeta(fields.Field(i).String()) + ").)*$"
					} else {
						op = "$nin"
					}
					break
				}

				if tmpField == "holding_item" {
					if op != "$eq" && op != "$ne" {
						continue
					}
					if op == "$eq" {
						if fields.Field(i).Bool() {
							bsonQueries = append(bsonQueries, bson.M{"pokemon.held_item": bson.M{"$ne": "None"}})
						} else {
							bsonQueries = append(bsonQueries, bson.M{"pokemon.held_item": bson.M{"$eq": "None"}})
						}
					} else {
						if fields.Field(i).Bool() {
							bsonQueries = append(bsonQueries, bson.M{"pokemon.held_item": bson.M{"$eq": "None"}})
						} else {
							bsonQueries = append(bsonQueries, bson.M{"pokemon.held_item": bson.M{"$ne": "None"}})
						}
					}
					continue
				} else if tmpField == "legal" {
					if op == "$eq" {
						if fields.Field(i).Bool() {
							bsonQueries = append(bsonQueries, bson.M{"pokemon.is_legal": true})
						} else {
							bsonQueries = append(bsonQueries, bson.M{"pokemon.is_legal": bson.M{"$in": []bool{true, false}}})
						}
					}
					continue
				}

				if wildcard {
					bsonQuery = bson.M{field: bson.M{"$regex": primitive.Regex{Pattern: op, Options: "i"}}}
				} else {
					bsonQuery = bson.M{field: bson.M{op: fields.Field(i).Interface()}}
				}
			}
		}
		if len(bsonQuery) == 0 {
			continue
		}

		bsonQueries = append(bsonQueries, bsonQuery)
	}

	if gpssPQ.MaxLevel != 0 && gpssPQ.MinLevel != 0 {
		bsonQueries = append(bsonQueries, bson.M{"pokemon.level": bson.M{"$gte": gpssPQ.MinLevel, "$lte": gpssPQ.MaxLevel}})
	}

	// Sorting
	sortField := ""

	switch gpssPQ.SortField {
	case "latest":
		sortField = "upload_date"
		break
	case "legality":
		sortField = "pokemon.is_legal"
		break
	case "popularity":
		sortField = "lifetime_downloads"
		break
	}

	direction := 0
	if *gpssPQ.SortDirection {
		direction = 1
	} else {
		direction = -1
	}

	sort = bson.M{sortField: direction}

	if gpssPQ.Mode == "" {
		return bson.M{}, sort, true, nil
	}

	// We only want approved pokemon showing up
	return bson.M{"approved": true, "deleted": false, fmt.Sprintf("$%s", gpssPQ.Mode): bsonQueries}, sort, true, gpssPQ.Generations
}

func ParseModerationQuery(w http.ResponseWriter, r *http.Request) (query bson.M, sort bson.M, ok bool) {

	return nil, nil, false
}

// FDecode decodes form data
func FDecode(w http.ResponseWriter, r *http.Request, v interface{}) (ok bool) {
	var err, rerr error

	if err = r.ParseForm(); err != nil {
		LogToSentry(err)
		rerr = fmt.Errorf("Error parsing %s parameters, invalid request", r.Method)
	} else {
		fdecoder := form.NewDecoder()
		switch r.Method {
		case http.MethodGet:
			err = fdecoder.Decode(v, r.Form)
		case http.MethodPost, http.MethodPut, http.MethodPatch:
			if strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
				dec := json.NewDecoder(r.Body)
				defer r.Body.Close()
				err = dec.Decode(v)
			} else {
				err = fdecoder.Decode(v, r.PostForm)
			}
		}
		if err != nil {
			rerr = fmt.Errorf("Error decoding %s request into required format (%T): validate request parameters", r.Method, v)
		}
	}

	if err != nil {
		_ = HttpError(w, r, http.StatusBadRequest, rerr, false, false, false)
	}

	return err == nil
}

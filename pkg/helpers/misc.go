package helpers

import (
	"encoding/json"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/FM1337/flagbrew-website-source-code/pkg/models"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func GenRandString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

// GetIP reads the request and gets the IP either from the CloudFlare header or by the remote address
func GetIP(r *http.Request, legacyKey string) string {
	ip := ""
	ipHeader := r.Header.Get("CF-Connecting-IP")
	legacyHeader := r.Header.Get("Flagbrew-V0-Legacy")
	if ipHeader == "" {
		if legacyHeader == legacyKey {
			ipHeader = r.Header.Get("Flagbrew-V0-Legacy-Ip")
			if ipHeader != "" {
				return ipHeader
			}
		}
		ipRemote, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			//LogToSentry(err)
			return r.RemoteAddr
		}
		ip = ipRemote
	} else {
		ip = ipHeader
	}
	return ip
}

// LoadSettings loads the settings into a map and returns it
func LoadSettings(settings []*models.Setting) map[string]*models.Setting {
	mapSettings := make(map[string]*models.Setting)
	if settings == nil {
		// use defaults
		for _, setting := range models.DefaultSettings {
			// Have to create a new setting model pointer rather than the one in the loop as all of them will be pointers to the last one.
			mapSettings[setting.MapKey] = &models.Setting{
				Name:           setting.Name,
				MapKey:         setting.MapKey,
				Description:    setting.Description,
				Type:           setting.Type,
				Value:          setting.Value,
				CanBeEmpty:     setting.CanBeEmpty,
				SystemVariable: setting.SystemVariable,
				CreatedBy:      setting.CreatedBy,
				CreatedDate:    setting.CreatedDate,
				ModifiedDate:   setting.ModifiedDate,
			}
		}
	} else {
		// use stored
		for _, setting := range settings {
			mapSettings[setting.MapKey] = setting
		}
	}
	return mapSettings
}

type badWordResult struct {
	IsBad           bool     `json:"is-bad"`
	BadWordList     []string `json:"bad-words-list"`
	BadWordsTotal   int      `json:"bad-words-total"`
	CensoredContent string   `json:"censored-content"`
}

// ProfanityLookup uses Neutrino Bad Work Lookup API to check the following content for profanity/obscene language:
// OT
// HT/NOT_OT
// Nickname
// ETC (aka whatever I think of)
func ProfanityLookup(accessKey string, lookupStrings []string) (successful bool, isBad bool) {
	result := badWordResult{}
	uri := "https://neutrinoapi.net/bad-word-filter"
	payload := url.Values{}
	payload.Set("user-id", "FMCore")
	payload.Set("api-key", accessKey)
	payload.Set("catalog", "strict")
	payload.Set("content", strings.Join(lookupStrings, " "))
	res, err := http.DefaultClient.PostForm(uri, payload)
	if err != nil {
		LogToSentry(err)
		return false, false
	}

	defer res.Body.Close()
	err = json.NewDecoder(res.Body).Decode(&result)
	if err != nil {
		LogToSentry(err)
		return false, false
	}
	// If there is
	if result.IsBad {
		return true, true
	}

	return true, false
}

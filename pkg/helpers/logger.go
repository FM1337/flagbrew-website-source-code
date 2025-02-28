package helpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/FM1337/flagbrew-website-source-code/pkg/models"
	"github.com/getsentry/sentry-go"
)

type SentryEventLogContext struct {
	IP        string                 `json:"ip,omitempty"`
	Path      string                 `json:"path,omitempty"`
	ExtraInfo map[string]interface{} `json:"extra_info"`
}

func GenerateSentryEventLogContext(keys []string, values []interface{}) (bool, SentryEventLogContext) {
	context := SentryEventLogContext{
		ExtraInfo: make(map[string]interface{}),
	}
	if len(keys) != len(values) {
		LogToSentry(fmt.Errorf("invalid amount of arguments provided, got %d keys and %d values", len(keys), len(values)))
		return false, context
	}

	for i, val := range values {

		switch keys[i] {
		case "ip":
			context.IP = val.(string)
		case "path":
			context.Path = val.(string)
		default:
			context.ExtraInfo[keys[i]] = val
		}
	}

	return true, context
}

func InitSentry(sentryDSN, environment string, debug bool) {
	err := sentry.Init(sentry.ClientOptions{
		Dsn:         sentryDSN,
		Environment: environment,
		Debug:       debug,
	})

	if err != nil {
		panic(err)
	}

}

func LogToSentry(err error) {
	sentry.CaptureException(err)
}

func LogToSentryWithContext(level sentry.Level, message string, context SentryEventLogContext) {
	event := sentry.NewEvent()
	event.Contexts = map[string]interface{}{"context": context}
	event.Level = level
	event.Message = message
	sentry.CaptureEvent(event)
}

type discordWebhookMessage struct {
	Embeds []discordEmbed `json:"embeds"`
}

type discordEmbed struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"url"`
	Color       int    `json:"color"`
	Timestamp   string `json:"timestamp"`
	Footer      struct {
		IconURL string `json:"icon_url"`
		Text    string `json:"text"`
	} `json:"footer"`
	Author struct {
		Name    string `json:"name"`
		URL     string `json:"url"`
		IconURL string `json:"icon_url"`
	} `json:"author"`
	Thumbnail struct {
		Url string `json:"url"`
	} `json:"thumbnail"`
	Fields []discordEmbedField `json:"fields"`
}

type discordEmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

func DiscordPostGPSS(uploadType, webhook, siteURL string, data interface{}) {
	embedObject := discordEmbed{}
	if uploadType == "individual" {
		pkmn := data.(models.GPSSPokemon)
		embedColor := 0
		if pkmn.Pokemon.IsLegal {
			embedColor = 0x10871e
		} else {
			embedColor = 0xbd3026
		}

		embedObject = discordEmbed{
			Title:       "GPSS Pokemon Uploaded",
			Description: "A new Pokemon has been uploaded to GPSS",
			URL:         fmt.Sprintf("%s/gpss/%s", siteURL, pkmn.DownloadCode),
			Color:       embedColor,
			Timestamp:   time.Now().Format(time.RFC3339),
		}
		embedObject.Footer.Text = fmt.Sprintf("Download Code: %s", pkmn.DownloadCode)
		embedObject.Footer.IconURL = pkmn.Pokemon.Sprites.Species
		embedObject.Thumbnail.Url = pkmn.Pokemon.Sprites.Species

		statFieldValue := ""
		for i, stat := range pkmn.Pokemon.Stats {
			tmpField := fmt.Sprintf("**%s**: %d, %d, %s", stat.Name, stat.IV, stat.EV, stat.Total)
			if i+1 != len(pkmn.Pokemon.Stats) {
				tmpField = fmt.Sprintf("%s\n", tmpField)
			}
			statFieldValue = fmt.Sprintf("%s%s", statFieldValue, tmpField)
		}

		genderIcon := "⚲"
		switch pkmn.Pokemon.Gender {
		case "F":
			genderIcon = "♀"
		case "M":
			genderIcon = "♂"
		}

		infoField := discordEmbedField{
			Name:   "Info",
			Inline: true,
			Value:  fmt.Sprintf("**Species**: %s\n**Level**: %d\n**Gender**: %s\n**Nickname**: %s\n**OT/ID**: %s/%d\n**Nature**: %s\n**Ability**: %s", pkmn.Pokemon.Species, pkmn.Pokemon.Level, genderIcon, pkmn.Pokemon.Nickname, pkmn.Pokemon.Ot, pkmn.Pokemon.Tid, pkmn.Pokemon.Nature, pkmn.Pokemon.Ability),
		}
		statsField := discordEmbedField{
			Name:   "Stats (IV | EV | Total)",
			Inline: true,
			Value:  statFieldValue,
		}
		lineBreakField := discordEmbedField{
			Name:  "\u200B",
			Value: "\u200B",
		}
		embedObject.Fields = append(embedObject.Fields, infoField, statsField, lineBreakField)
		for i, move := range pkmn.Pokemon.Moves {
			if move.Name == "None" {
				continue
			}
			tmpField := discordEmbedField{
				Name:   fmt.Sprintf("Move %d", i+1),
				Inline: true,
				Value:  fmt.Sprintf("**Name**: [%s](%s)\n**Type**: %s\n**PP**: %d\n**PP Ups**: %d", move.Name, serebiiMoveLink(pkmn.Generation, move.Name), move.Type, move.PP, move.PPUps),
			}
			if i == 2 {
				embedObject.Fields = append(embedObject.Fields, lineBreakField)
			}
			embedObject.Fields = append(embedObject.Fields, tmpField)
		}
	} else {
		bundle := data.(*models.GPSSBundlePokemon)
		embedColor := 0
		if bundle.IsLegal {
			embedColor = 0x10871e
		} else {
			embedColor = 0xbd3026
		}
		embedObject = discordEmbed{
			Title:       "GPSS Bundle Uploaded",
			Description: "A new bundle has been uploaded to GPSS",
			Color:       embedColor,
			Timestamp:   time.Now().Format(time.RFC3339),
		}
		embedObject.Footer.Text = fmt.Sprintf("Download Code: %s", bundle.DownloadCode)

		for i, pk := range bundle.Pokemons {
			legality := "✅"
			if !pk.Legality {
				legality = "❌"
			}
			tmpField := discordEmbedField{
				Name:   fmt.Sprintf("PKMN %d", i+1),
				Inline: true,
				Value:  fmt.Sprintf("**Download Code**: [%s](%s/gpss/%s)\n**Legality**: %s\n**Generation**: %d", bundle.DownloadCodes[i], siteURL, bundle.DownloadCodes[i], legality, pk.Generation),
			}
			embedObject.Fields = append(embedObject.Fields, tmpField)
		}
	}

	embedObject.Author.Name = "GPSS Upload"
	embedObject.Author.IconURL = "https://cdn.discordapp.com/attachments/807696957056221235/827331872014598144/image0.png"
	embedObject.Author.URL = fmt.Sprintf("%s/gpss", siteURL)

	sendEmbed := &discordWebhookMessage{
		Embeds: []discordEmbed{embedObject},
	}
	jsonValue, _ := json.Marshal(sendEmbed)
	http.Post(webhook, "application/json", bytes.NewBuffer(jsonValue))
}

func serebiiMoveLink(generation int, moveName string) string {
	dex := ""
	switch generation {
	case 1:
		dex = "-rby"
	case 2:
		dex = "-gs"
	case 4:
		dex = "-dp"
	case 5:
		dex = "-bw"
	case 6:
		dex = "-xy"
	case 7:
		dex = "-sm"
	case 8:
		dex = "-swsh"
	case 9:
		dex = "-sv"
	}

	move := strings.ReplaceAll(strings.ToLower(moveName), " ", "")

	return fmt.Sprintf("https://serebii.net/attackdex%s/%s.shtml", dex, move)
}

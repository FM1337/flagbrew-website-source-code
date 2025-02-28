package api

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"

	"github.com/FM1337/flagbrew-website-source-code/pkg/helpers"
	"github.com/FM1337/flagbrew-website-source-code/pkg/models"
)

type GPSSAPI struct{}

func NewGPSSAPI() *GPSSAPI {
	return &GPSSAPI{}
}

func (*GPSSAPI) GetPokemonInfo(pokemonFile []byte, formData map[string]string) (pokemon *models.Pokemon, err error) {
	// endpoint is api/info
	endpoint := "/info"
	data, success, err := helpers.CoreAPIFile(pokemonFile, formData, endpoint)
	if err != nil || !success {
		return pokemon, err
	}
	if helpers.IsArugmentError(errors.New(strings.Split(string(data), "\n")[0])) {
		err = errors.New("there is an error in your provided information")
		return pokemon, err
	}
	err = json.Unmarshal(data, &pokemon)
	return pokemon, err
}

func (*GPSSAPI) GetLegalityInfo(pokemon []byte, formData map[string]string) (legality *models.LegalityInfo, err error) {
	endpoint := "/legality/check"

	data, _, err := helpers.CoreAPIFile(pokemon, formData, endpoint)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &legality)
	return legality, err
}

func (*GPSSAPI) AutoLegalize(pokemon []byte, formData map[string]string) (legalize *models.AutoLegalize, err error) {
	endpoint := "/legality/legalize"
	data, _, err := helpers.CoreAPIFile(pokemon, formData, endpoint)
	if err != nil {
		return nil, err
	}

	if strings.Contains(string(data), "this pokemon is already legal!") {
		return &models.AutoLegalize{
			Ran:     false,
			Pokemon: base64.StdEncoding.EncodeToString(pokemon), // base64 encoded pokemon
			Legal:   true,
			Report:  []string{"legal!"},
		}, nil
	}

	err = json.Unmarshal(data, &legalize)
	if err != nil {
		return nil, err
	}

	legalize.Ran = true
	return legalize, nil
}

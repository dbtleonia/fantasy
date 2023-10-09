package yahoo

import (
	"encoding/json"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/endpoints"
)

func ReadConfig(filename string) (*oauth2.Config, error) {
	b, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var conf oauth2.Config
	if err := json.Unmarshal(b, &conf); err != nil {
		return nil, err
	}
	conf.RedirectURL = "oob"
	conf.Endpoint = endpoints.Yahoo
	return &conf, nil
}

func ReadToken(filename string) (*oauth2.Token, error) {
	b, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var tok oauth2.Token
	if err = json.Unmarshal(b, &tok); err != nil {
		return nil, err
	}
	return &tok, nil
}

func WriteToken(tok *oauth2.Token, filename string) error {
	b, err := json.MarshalIndent(tok, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, []byte("\n")...)
	return os.WriteFile(filename, b, 0600)
}

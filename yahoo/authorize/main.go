package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path"

	"github.com/dbtleonia/fantasy/yahoo"
	"golang.org/x/oauth2"
)

const (
	secretsFile = "client_secrets.json"
	tokenFile   = "token.json"
)

func main() {
	home, _ := os.UserHomeDir()

	//
	// Mostly copied from https://pkg.go.dev/golang.org/x/oauth2#example-Config
	//
	ctx := context.Background()
	conf, err := yahoo.ReadConfig(path.Join(home, secretsFile))
	if err != nil {
		log.Fatal(err)
	}

	// Redirect user to consent page to ask for permission
	// for the scopes specified above.
	url := conf.AuthCodeURL("state", oauth2.AccessTypeOffline)
	fmt.Printf("Visit the URL for the auth dialog: %v\n", url)
	fmt.Printf("\n")
	fmt.Printf("Enter code: ")

	// Use the authorization code that is pushed to the redirect
	// URL. Exchange will do the handshake to retrieve the
	// initial access token. The HTTP Client returned by
	// conf.Client will refresh the token as necessary.
	var code string
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatal(err)
	}
	tok, err := conf.Exchange(ctx, code)
	if err != nil {
		log.Fatal(err)
	}

	filename := path.Join(home, tokenFile)
	if err := yahoo.WriteToken(tok, filename); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\n")
	fmt.Printf("SUCCESS: Wrote %s\n", filename)
}

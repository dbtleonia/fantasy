package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/dbtleonia/fantasy/yahoo"
	"golang.org/x/oauth2"
)

func leaguePlayersAvailable(client *http.Client, league, suffix string) error {
	// TODO: Don't hardcode limit.
	for start := 0; start < 950; start += 25 {
		u := fmt.Sprintf("/league/%s/players;status=A;start=%d;count=25/stats", league, start)
		f := fmt.Sprintf("league_players_available_%s_%04d.json", suffix, start)
		if err := getToFile(client, u, f); err != nil {
			return err
		}
	}
	return nil
}

func leagueScoreboard(client *http.Client, league string, week int) error {
	u := fmt.Sprintf("/league/%s/scoreboard;week=%d", league, week)
	f := fmt.Sprintf("league_scoreboard_week%02d.json", week)
	return getToFile(client, u, f)
}

func leagueStandings(client *http.Client, league, suffix string) error {
	u := fmt.Sprintf("/league/%s/standings", league)
	f := fmt.Sprintf("league_standings_%s.json", suffix)
	if err := getToFile(client, u, f); err != nil {
		return err
	}

	u2 := fmt.Sprintf("/league/%s/settings", league)
	f2 := fmt.Sprintf("league_settings_%s.json", suffix)
	return getToFile(client, u2, f2)
}

func teamMatchups(client *http.Client, league string) error {
	for t := 1; t <= 12; t++ {
		u := fmt.Sprintf("/team/%s.t.%d/matchups", league, t)
		f := fmt.Sprintf("team_matchups_%02d.json", t)
		if err := getToFile(client, u, f); err != nil {
			return err
		}
	}
	return nil
}

func teamRosters(client *http.Client, league string, week int) error {
	for t := 1; t <= 12; t++ {
		u := fmt.Sprintf("/team/%s.t.%d/roster;week=%d/players/stats", league, t, week)
		f := fmt.Sprintf("team_rosters_week%02d_%02d.json", week, t)
		if err := getToFile(client, u, f); err != nil {
			return err
		}
	}
	return nil
}

func getToFile(client *http.Client, uriPath, file string) error {
	resp, err := client.Get(fullURI(uriPath))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	home, _ := os.UserHomeDir()
	filename := path.Join(home, file)
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	_, err = io.Copy(f, resp.Body)
	if _, err0 := f.Write([]byte("\n")); err0 != nil && err == nil {
		err = err0
	}
	if err1 := f.Close(); err1 != nil && err == nil {
		err = err1
	}
	if err == nil {
		log.Printf("Wrote %s\n", filename)
	}
	return err
}

func getToStdout(client *http.Client, uri string) error {
	if !strings.HasPrefix(uri, "https://") {
		uri = fullURI(uri)
	}
	resp, err := client.Get(uri)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = io.Copy(os.Stdout, resp.Body)
	return err
}

func fullURI(uriPath string) string {
	return fmt.Sprintf("https://fantasysports.yahooapis.com/fantasy/v2%s?format=json_f", uriPath)
}

const (
	leagueFile  = "league.txt"
	secretsFile = "client_secrets.json"
	tokenFile   = "token.json"
)

func main() {
	ctx := context.Background()

	home, _ := os.UserHomeDir()

	b, err := os.ReadFile(path.Join(home, leagueFile))
	if err != nil {
		log.Fatal(err)
	}
	league := string(bytes.TrimSpace(b))

	conf, err := yahoo.ReadConfig(path.Join(home, secretsFile))
	if err != nil {
		log.Fatal(err)
	}

	tok, err := yahoo.ReadToken(path.Join(home, tokenFile))
	if err != nil {
		log.Fatal(err)
	}

	client := oauth2.NewClient(ctx, NewCachingTokenSource(path.Join(home, tokenFile), conf, tok))

	if len(os.Args) == 1 {
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Printf("Enter URI: ")
		for scanner.Scan() {
			uri := scanner.Text()
			if err := getToStdout(client, uri); err != nil {
				log.Println(err)
			}
			fmt.Printf("\n\n")
			fmt.Printf("Enter URI: ")
		}
		return
	}

	switch os.Args[1] {
	case "a":
		if len(os.Args) != 3 {
			log.Fatal("gimme suffix")
		}
		err = leaguePlayersAvailable(client, league, os.Args[2])
	case "b":
		if len(os.Args) != 3 {
			log.Fatal("gimme week")
		}
		week, err := strconv.Atoi(os.Args[2])
		if err != nil {
			log.Fatal(err)
		}
		err = leagueScoreboard(client, league, week)
	case "m":
		err = teamMatchups(client, league)
	case "r":
		if len(os.Args) != 3 {
			log.Fatal("gimme week")
		}
		week, err := strconv.Atoi(os.Args[2])
		if err != nil {
			log.Fatal(err)
		}
		err = teamRosters(client, league, week)
	case "s":
		if len(os.Args) != 3 {
			log.Fatal("gimme suffix")
		}
		err = leagueStandings(client, league, os.Args[2])
	default:
		err = getToStdout(client, os.Args[1])
	}
	if err != nil {
		log.Fatal(err)
	}
}

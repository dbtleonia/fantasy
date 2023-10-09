// From https://github.com/golang/oauth2/issues/84#issuecomment-1542490270

package main

import (
	"context"
	"log"

	"github.com/dbtleonia/fantasy/yahoo"
	"golang.org/x/oauth2"
)

type cachingTokenSource struct {
	base     oauth2.TokenSource
	filename string
}

func (c *cachingTokenSource) saveToken(tok *oauth2.Token) error {
	return yahoo.WriteToken(tok, c.filename)
}

func (c *cachingTokenSource) loadToken() (*oauth2.Token, error) {
	return yahoo.ReadToken(c.filename)
}

func (c *cachingTokenSource) Token() (tok *oauth2.Token, err error) {
	tok, _ = c.loadToken()
	if tok != nil && tok.Valid() {
		return tok, nil
	}

	if tok, err = c.base.Token(); err != nil {
		return nil, err
	}

	err2 := c.saveToken(tok)
	if err2 != nil {
		log.Print("cache token:", err2) // or return it
	}

	return tok, err
}

func NewCachingTokenSource(filename string, config *oauth2.Config, tok *oauth2.Token) oauth2.TokenSource {
	orig := config.TokenSource(context.Background(), tok)
	return oauth2.ReuseTokenSource(nil, &cachingTokenSource{
		filename: filename,
		base:     orig,
	})
}

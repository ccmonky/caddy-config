package http

import "github.com/caddyserver/caddy/v2"

type HTTP struct {
	*Clients
	*RequestBuilders
	*Handlers
	*MatcherSets
}

func (http *HTTP) Provision(caddy.Context) error {
	return nil
}

func (http *HTTP) Validate() error {
	return nil
}

type Clients struct{}

func (cs *Clients) Provision(caddy.Context) error {
	return nil
}

func (cs *Clients) Validate() error {
	return nil
}

type RequestBuilders struct{}

func (bs *RequestBuilders) Provision(caddy.Context) error {
	return nil
}
func (bs *RequestBuilders) Validate() error {
	return nil
}

type Handlers struct{}

func (hs *Handlers) Provision(caddy.Context) error {
	return nil
}
func (hs *Handlers) Validate() error {
	return nil
}

type MatcherSets struct{}

func (ms *MatcherSets) Provision(caddy.Context) error {
	return nil
}
func (ms *MatcherSets) Validate() error {
	return nil
}

package dynconf

import "github.com/caddyserver/caddy/v2"

type HTTP struct {
	URL      string         `json:"url"`
	Method   string         `json:"method,omitempty"`
	Interval caddy.Duration `json:"interval,omitempty"`
}

func (h HTTP) ID() string {
	return "config.extension.dynconf.http"
}

package nacos

import (
	"github.com/caddyserver/caddy/v2"
	"github.com/ccmonky/caddy-config/dynconf"
	"github.com/ccmonky/typemap"
)

// Not implemnt...
type Nacos struct {
	ServerConfig string      `json:"server_config"`
	ClientConfig string      `json:"client_config"`
	Datas        []NacosData `json:"datas"`
}

type NacosData struct {
	Group     string                          `json:"group"`
	DataId    string                          `json:"data_id"`
	Callbacks []typemap.Ref[dynconf.Callback] `json:"callbacks"`
}

func (n Nacos) ID() string {
	return "config.ext.dynconf.listeners.nacos"
}

func (n *Nacos) Provision(ctx caddy.Context) error {
	return nil
}

// Interface guard
var (
	_ caddy.Provisioner = (*Nacos)(nil)
)

package mock

import (
	"context"
	"encoding/json"

	"github.com/caddyserver/caddy/v2"
	"github.com/pkg/errors"

	"github.com/ccmonky/pkg/mock"
	"github.com/ccmonky/typemap"
)

// IMatcher 由所有config.mock.matchers.xxx模块实现，供Matchers模块注册资源使用
// 设计考量：不一定每个模块自身都会实现mock.Matcher，因此，实现IMatcher更具通用性
type IMatcher interface {
	Matcher() mock.Matcher
}

func init() {
	caddy.RegisterModule(Matchers{})
}

// Clients 定义HTTP客户端，便于多个包共享连接池
type Matchers struct {
	Matchers []Matcher `json:"matchers,omitempty"`
}

// Client define a http client config
type Matcher struct {
	Name      string          `json:"name"`
	ConfigRaw json.RawMessage `json:"config" caddy:"namespace=config.mock.matchers inline_key=matcher"`
}

// ID 模块ID
func (Matchers) ID() string {
	return "config.mock.matchers"
}

// CaddyModule returns the Caddy module information.
func (c Matchers) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  caddy.ModuleID(c.ID()),
		New: func() caddy.Module { return new(Matchers) },
	}
}

// Provision 实现Provisioner
func (c *Matchers) Provision(ctx caddy.Context) error {
	sentinel := map[string]struct{}{}
	for _, matcherConf := range c.Matchers {
		name := matcherConf.Name
		if name == "" {
			return errors.New("config.mock.matchers encounter empty name")
		}
		if _, ok := sentinel[name]; ok {
			return errors.Errorf("mock matcher %s repeated", name)
		}
		value, err := ctx.LoadModule(&matcherConf, "ConfigRaw")
		if err != nil {
			return errors.Wrapf(err, "load %s %s failed", c.ID(), name)
		}
		matcher, ok := value.(IMatcher)
		if !ok {
			return errors.Errorf("get mock matcher %s from registry not implement IMatcher", name)
		}
		err = typemap.Register[mock.Matcher](ctx, name, matcher.Matcher())
		if err != nil {
			return errors.WithMessagef(err, "register mock.Matcher %s failed", name)
		}
		sentinel[name] = struct{}{}
	}
	return nil
}

// Validate 实现Validator
func (c Matchers) Validate() error {
	for _, matcherConf := range c.Matchers {
		v, err := typemap.Get[mock.Matcher](context.Background(), matcherConf.Name)
		if err != nil {
			return errors.WithMessagef(err, "get resource %s instance %s failed", typemap.GetTypeIdString[mock.Matcher](), matcherConf.Name)
		}
		if v == nil {
			return errors.Errorf("config.mock.matchers %s is nil pointer", matcherConf.Name)
		}
	}
	return nil
}

// Produces 记录资源和模块生产关系
func (c Matchers) Produces() []string {
	return []string{
		typemap.GetTypeIdString[mock.Matcher](),
	}
}

// GetResourceInstanceNames 获取资源实例名称
// func (c Matchers) GetResourceInstanceNames() (map[string][]string, error) {
// 	//Clients 资源实例
// 	matcherProduces := []string{mock.MetaOfMatcher.Name()}
// 	rInstanceNames, err := config.GetResourceInstanceNames(matcherProduces, c.Matchers, "name")
// 	if err != nil {
// 		return nil, err
// 	}
// 	return rInstanceNames, nil
// }

// Interface guard
var (
	_ caddy.Validator   = (*Matchers)(nil)
	_ caddy.Provisioner = (*Matchers)(nil)
	//_ modules.Producer  = (*Matchers)(nil)
)

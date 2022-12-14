package caddyconfig

import (
	"encoding/json"
	"fmt"

	"github.com/caddyserver/caddy/v2"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/ccmonky/caddy-config/http"
	"github.com/ccmonky/caddy-config/logging"
	"github.com/ccmonky/caddy-config/mock"
	"github.com/ccmonky/caddy-config/pool"
	"github.com/ccmonky/caddy-config/trace"
)

func init() {
	caddy.RegisterModule(Config{})
}

// Config 配置平台
type Config struct {
	// PreApps 设定前置Apps
	PreApps []string `json:"pre_apps"`

	// Registries 资源注册表?
	//Registries *registry.Registries

	// Logging 定义应用依赖的三方库使用的logger资源
	Logging *logging.Logging `json:"logging,omitempty"`

	// Tracers 配置兼容opentracing的Tracers
	Tracers *trace.Tracers `json:"tracers,omitemtpy"`

	// EigenkeyRaw 特征键提取器
	EigenkeyRaw map[string]json.RawMessage `json:"eigenkey,omitempty"`

	StoreRaw map[string]json.RawMessage `json:"store,omitempty"`

	// Mock 配置mock
	Mock *mock.Mock `json:"mock,omitempty"`

	// HTTP 配置http clients和handlers
	HTTP *http.HTTP `json:"http,omitempty"`

	// Pool 通用对象池
	Pool *pool.Pool `json:"pool,omitempty"`

	ExtensionRaw map[string]json.RawMessage `json:"extension,omitempty"`

	readyMods map[string]Ready
	ctx       caddy.Context
	logger    *zap.Logger
}

// Ready 用于执行就绪检查，有些tproxy插件在Provision和Validate阶段无法做引用资源有效性检查
type Ready interface {
	Ready() error
}

// CaddyModule returns the Caddy module information.
func (Config) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "config",
		New: func() caddy.Module { return new(Config) },
	}
}

// Provision sets up the app.
func (c *Config) Provision(ctx caddy.Context) error {
	c.ctx = ctx
	c.logger = ctx.Logger(c)
	for _, appName := range c.PreApps {
		if appName == "config" {
			continue
		}
		_, err := ctx.App(appName) // NOTE: ensure `appName` App already provisioned
		if err != nil {
			return errors.Wrapf(err, "config rely on %s App failed", appName)
		}
	}

	// NOTE: 由于CodecCacher需要设定newInstanceFunc和loadInstanceFunc，因此放入具体模块初始化阶段定义Meta和Cacher！
	// if c.Registries != nil {
	// 	err := c.Registries.Provision(ctx)
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	if c.Logging != nil {
		err := c.Logging.Provision(ctx)
		if err != nil {
			return err
		}
	}

	if c.Tracers != nil {
		err := c.Tracers.Provision(ctx)
		if err != nil {
			return err
		}
	}

	c.readyMods = make(map[string]Ready)

	for typ, rawMsg := range c.EigenkeyRaw {
		_, err := ctx.LoadModuleByID("config.eigenkey."+typ, rawMsg)
		if err != nil {
			return fmt.Errorf("loading eigenkey config module in position %s: %v", typ, err)
		}
	}
	c.EigenkeyRaw = nil // allow GC to deallocate

	for driver, rawMsg := range c.StoreRaw {
		_, err := ctx.LoadModuleByID("config.store."+driver, rawMsg)
		if err != nil {
			return fmt.Errorf("loading store config module in position %s: %v", driver, err)
		}
	}
	c.StoreRaw = nil // allow GC to deallocate

	if c.Mock != nil {
		if c.Mock.Matchers != nil {
			err := c.Mock.Matchers.Provision(ctx)
			if err != nil {
				return err
			}
		}
	}

	if c.HTTP != nil {
		if c.HTTP.Clients != nil {
			err := c.HTTP.Clients.Provision(ctx)
			if err != nil {
				return err
			}
		}
		if c.HTTP.RequestBuilders != nil {
			err := c.HTTP.RequestBuilders.Provision(ctx)
			if err != nil {
				return err
			}
		}
	}

	if c.Pool != nil {
		err := c.Pool.Provision(ctx)
		if err != nil {
			return err
		}
	}

	for name, rawMsg := range c.ExtensionRaw {
		_, err := ctx.LoadModuleByID("config.goapp."+name, rawMsg)
		if err != nil {
			return fmt.Errorf("loading goapp config module in position %s: %v", name, err)
		}
	}
	c.ExtensionRaw = nil // allow GC to deallocate

	// NOTE: Handlers 放在最后，因为要先让goapp注册HandlerFuncMap！
	if c.HTTP != nil {
		if c.HTTP.Handlers != nil {
			err := c.HTTP.Handlers.Provision(ctx)
			if err != nil {
				return err
			}
		}
		if c.HTTP.MatcherSets != nil {
			err := c.HTTP.MatcherSets.Provision(ctx)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Validate ensures the app's configuration is valid.
func (c *Config) Validate() error {
	// if c.Registries != nil {
	// 	err := c.Registries.Validate()
	// 	if err != nil {
	// 		return err
	// 	}
	// }
	if c.Logging != nil {
		err := c.Logging.Validate()
		if err != nil {
			return err
		}
	}
	if c.Pool != nil {
		err := c.Pool.Validate()
		if err != nil {
			return err
		}
	}
	if c.Mock != nil {
		return c.Mock.Validate()
	}
	if c.HTTP != nil {
		return c.HTTP.Validate()
	}
	return nil
}

// Start runs the app
func (c *Config) Start() error {
	// 0. NOTE: 这实际上是在做Validate操作，当前主要场景是sso需要验证jwtauth的实例是否合法！
	for name, readyMod := range c.readyMods {
		err := readyMod.Ready()
		if err != nil {
			return errors.Errorf("mod %s Ready check failed: %v", name, err)
		}
	}
	// 1. 注册BuiltinAPIRouter
	// NOTE: 为什么放这里？因此有些如sso、errorspace可能是在provision里才注册BuiltinAPIRouter，此处Provision已经全部执行完毕！
	return registerBuiltinAPIRouters()
}

// Stop gracefully shuts down the HTTP server.
func (c *Config) Stop() error {
	return nil
}

// Interface guard
var (
	_ caddy.App         = (*Config)(nil)
	_ caddy.Validator   = (*Config)(nil)
	_ caddy.Provisioner = (*Config)(nil)
)

package caddyconfig

import (
	"fmt"

	"github.com/caddyserver/caddy/v2"
)

func ProvisionPreApps(ctx caddy.Context, curApp string, preApps []string) error {
	for _, appName := range preApps {
		if appName == curApp {
			continue
		}
		_, err := ctx.App(appName) // NOTE: ensure `appName` App already provisioned
		if err != nil {
			return fmt.Errorf("%s provision pre app %s failed: %v", curApp, appName, err)
		}
	}
	return nil
}

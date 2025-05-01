package config

import "github.com/caarlos0/env/v11"

func ParseConfigFromEnv(appConfig *AppConfig) error {
	if err := env.Parse(appConfig); err != nil {
		return err
	}
	return nil
}

package config

import "github.com/BurntSushi/toml"

func loadConfig(path string) (Config, error) {
	var config Config
	_, err := toml.DecodeFile(path, &config)
	return config, err
}

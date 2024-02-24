package apocli

import (
	"apocli/pkg/anova"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/kirsle/configdir"
	"os"
	"path/filepath"
)

const (
	ConfigFolderName = "apocli"
	ConfigFileName   = "config.toml"
)

type Config struct {
	FirebaseRefreshToken string

	DefaultCookerID anova.CookerID
}

func DefaultConfig() *Config {
	return &Config{
		FirebaseRefreshToken: "",
		DefaultCookerID:      "",
	}
}

func ConfigFilePath() (string, error) {
	configPath := configdir.LocalConfig(ConfigFolderName)
	err := configdir.MakePath(configPath)
	if err != nil {
		return "", fmt.Errorf("unable to create config path \"%s\"", configPath)
	}

	return filepath.Join(configPath, ConfigFileName), nil
}

func (config *Config) Save() error {
	configFilePath, err := ConfigFilePath()
	if err != nil {
		return err
	}

	// Always create the file from scratch
	fh, err := os.Create(configFilePath)
	if err != nil {
		return fmt.Errorf("failed to create config file \"%s\": %+v", configFilePath, err)
	}
	defer fh.Close()

	encoder := toml.NewEncoder(fh)
	err = encoder.Encode(config)
	if err != nil {
		return fmt.Errorf("failed to write config file \"%s\": %+v", configFilePath, err)
	}

	return nil
}

func LoadConfig() (*Config, error) {
	configFilePath, err := ConfigFilePath()
	if err != nil {
		return nil, err
	}

	config := DefaultConfig()

	if _, err = os.Stat(configFilePath); os.IsNotExist(err) {
		fh, err := os.Create(configFilePath)
		if err != nil {
			return nil, fmt.Errorf("failed to create config file \"%s\": %+v", configFilePath, err)
		}
		defer fh.Close()

		encoder := toml.NewEncoder(fh)
		err = encoder.Encode(config)
		if err != nil {
			return nil, fmt.Errorf("failed to write config file \"%s\": %+v", configFilePath, err)
		}
	} else {
		fh, err := os.Open(configFilePath)
		if err != nil {
			return nil, fmt.Errorf("failed to open config file \"%s\": %+v", configFilePath, err)
		}
		defer fh.Close()

		decoder := toml.NewDecoder(fh)
		_, err = decoder.Decode(config)
		if err != nil {
			return nil, fmt.Errorf("failed to decode config file \"%s\": %+v", configFilePath, err)
		}
	}

	return config, nil
}

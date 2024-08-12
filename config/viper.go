package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	Debug    bool           `toml:"debug" mapstructure:"debug" json:"debug" yaml:"debug"`
	API      apiConfig      `toml:"api" mapstructure:"api" json:"api" yaml:"api"`
	Fetcher  fetcherConfig  `toml:"fetcher" mapstructure:"fetcher" json:"fetcher" yaml:"fetcher"`
	Log      logConfig      `toml:"log" mapstructure:"log" json:"log" yaml:"log"`
	Source   sourceConfigs  `toml:"source" mapstructure:"source" json:"source" yaml:"source"`
	Storage  storageConfigs `toml:"storage" mapstructure:"storage" json:"storage" yaml:"storage"`
	Telegram telegramConfig `toml:"telegram" mapstructure:"telegram" json:"telegram" yaml:"telegram"`
	Database databaseConfig `toml:"database" mapstructure:"database" json:"database" yaml:"database"`
}

type fetcherConfig struct {
	MaxConcurrent int `toml:"max_concurrent" mapstructure:"max_concurrent" json:"max_concurrent" yaml:"max_concurrent"`
	Limit         int `toml:"limit" mapstructure:"limit" json:"limit" yaml:"limit"`
}

var Cfg *Config

func InitConfig() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.SetConfigType("toml")

	viper.SetDefault("api.enable", false)
	viper.SetDefault("api.address", "0.0.0.0:39080")
	viper.SetDefault("api.auth", false)

	viper.SetDefault("fetcher.max_concurrent", 5)
	viper.SetDefault("fetcher.limit", 30)

	viper.SetDefault("log.level", "TRACE")
	viper.SetDefault("log.file_path", "logs/ManyACG.log")
	viper.SetDefault("log.backup_num", 7)

	viper.SetDefault("source.pixiv.enable", false)
	viper.SetDefault("source.twitter.enable", false)
	viper.SetDefault("source.twitter.fx_twitter_domain", "fxtwitter.com")
	viper.SetDefault("source.bilibili.enable", false)
	viper.SetDefault("source.danbooru.enable", false)
	viper.SetDefault("source.kemono.enable", false)

	viper.SetDefault("storage.cache_dir", "./cache")
	viper.SetDefault("storage.cache_ttl", 86400)
	viper.SetDefault("storage.default", "local")
	viper.SetDefault("storage.local.enable", true)
	viper.SetDefault("storage.local.path", "./manyacg")

	viper.SetDefault("telegram.sleep", 1)
	viper.SetDefault("telegram.api_url", "https://api.telegram.org")

	viper.SetDefault("Database.databse", "manyacg")

	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("error when reading config: %s\n", err)
		os.Exit(1)
	}
	Cfg = &Config{}
	if err := viper.Unmarshal(Cfg); err != nil {
		fmt.Printf("error when unmarshal config: %s\n", err)
		os.Exit(1)
	}

	if len(Cfg.Telegram.Admins) == 0 {
		fmt.Println("please set at least one admin in config file (telegram.admins)")
		os.Exit(1)
	}
}

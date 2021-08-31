package config

import (
	"fmt"
	"github.com/spf13/viper"
)

type Config struct {
	TelegramApiKey  string
	NotValidTitles  []string
	NotValidSizes   []string
	MinimumDiscount float64
	Categories      []Category
}

type Category struct {
	Channel string `json:"channel"`
	URL             string
}

func LoadConfigs() *Config {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	var category []Category
	err = viper.UnmarshalKey("categories", &category)
	if err != nil {
		fmt.Println(err)
	}
	return &Config{
		TelegramApiKey:  viper.GetString("telegram_api_key"),
		NotValidTitles:  viper.GetStringSlice("not_valid_titles"),
		NotValidSizes:   viper.GetStringSlice("not_valid_sizes"),
		MinimumDiscount: viper.GetFloat64("minimum_discount"),
		Categories:      category,
	}
}

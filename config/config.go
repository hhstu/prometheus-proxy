package config

import (
	"fmt"
	"github.com/spf13/viper"
	"path"
	"runtime"
)

var AppConfig Config

func init() {
	_, config, _, _ := runtime.Caller(0)
	viper.AddConfigPath(path.Dir(config))
	viper.AddConfigPath("./config")
	viper.AddConfigPath("./")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %w \n", err))
	}
	if err := viper.Unmarshal(&AppConfig); err != nil {
		panic(fmt.Errorf("unmarshal error: %w \n", err))
	}

}

type Config struct {
	Webserver struct {
		Port string `json:"port"`
		Mode string `json:"mode"`
	} `json:"webserver"`
	Log struct {
		Level    string `json:"level"`
		Encoding string `json:"encoding"`
	} `json:"log"`
	PrometheusUrl string   `json:"prometheus_url"`
	Headers       []string `json:"headers"`
	Params        []string `json:"params"`
}

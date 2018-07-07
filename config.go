package main

import (
	"github.com/spf13/viper"
)

func Config() (string, bool, []string, string) {
	viper.SetConfigName("aisetl")
	viper.SetConfigName(".aisetl")
	viper.AddConfigPath("/etc/")
	viper.AddConfigPath("/usr/local/etc/")
	viper.AddConfigPath("$HOME/")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	viper.WatchConfig()

	f := viper.GetStringSlice("forward")

	forward := true
	if len(f) == 0 {
		forward = false
	}

	return viper.GetString("redis"), forward, f, viper.GetString("listen")
}

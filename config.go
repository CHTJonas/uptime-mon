package main

import (
	"fmt"
	"os"
	"sync/atomic"

	"github.com/spf13/viper"
)

var config *atomic.Value

type Config struct {
	slackWebhook string
	tests        []*Test
}

func init() {
	fmt.Printf("Uptime Monitor %s started\n", version)
	config = new(atomic.Value)
	config.Store(loadConfig())
}

func getConfig() *Config {
	return config.Load().(*Config)
}

func readConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/uptime-mon/")
	viper.AddConfigPath("$HOME/.config/uptime-mon/")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("Failed to read config file:", err)
		os.Exit(125)
	}
}

func loadConfig() *Config {
	readConfig()
	settingsInConfig := viper.GetStringMapString("settings")
	testsInConfig := viper.Get("tests").([]interface{})
	size := len(testsInConfig)
	c := &Config{}
	c.slackWebhook = settingsInConfig["slack-webhook"]
	c.tests = make([]*Test, size)
	for i, t := range testsInConfig {
		c.tests[i] = NewTest(t.(map[interface{}]interface{}))
	}
	fmt.Println("Found", size, "tests in config file")
	return c
}

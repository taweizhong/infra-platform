package main

import (
	"log"
	"os"

	"infra-platform/sdk/configsdk"
)

type AppConfig struct {
	Database struct {
		DSN string `toml:"dsn"`
	} `toml:"database"`
}

func main() {
	client, err := configsdk.New(
		configsdk.WithAgent("http://config-agent:8080"),
		configsdk.WithApp("user-svc"),
		configsdk.WithEnv("prod"),
		configsdk.WithCluster("cn-east-1"),
		configsdk.WithWatch(true),
	)
	if err != nil {
		log.Printf("sdk init failed: %v", err)
		os.Exit(1)
	}
	defer client.Close()

	var cfg AppConfig
	if err := client.Config().Unmarshal(&cfg); err != nil {
		log.Printf("unmarshal failed: %v", err)
		os.Exit(1)
	}
}

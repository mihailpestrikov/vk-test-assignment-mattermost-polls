package main

import (
	"vk-test-assignment-mattermost-polls/pkg/config"
	"vk-test-assignment-mattermost-polls/pkg/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		println("Failed to load config:", err.Error())
		return
	}

	logger.Setup(cfg.Logger)
}

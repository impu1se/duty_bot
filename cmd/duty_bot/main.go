package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/impu1se/duty_bot/configs"
	"github.com/impu1se/duty_bot/internal/botapi"
	"github.com/impu1se/duty_bot/internal/duty"
	"github.com/impu1se/duty_bot/internal/storage"
	"go.uber.org/zap"
)

func main() {

	config := configs.NewConfig()

	botApi, err := botapi.NewBotApi(config)
	if err != nil {
		log.Fatalf("can't get new bot api, reason: %v", err)
	}

	if config.Tls {
		go http.ListenAndServeTLS(":"+config.Port, config.CertFile, config.KeyFile, nil)
	} else {
		go http.ListenAndServe(":"+config.Port, nil)
	}

	logger := zap.NewExample()
	system := storage.NewLoader(logger)
	gifBot := duty.NewDutyBot(config, botApi.ListenForWebhook("/"+botApi.Token), system, logger, *botApi, context.Background())

	fmt.Printf("Start server on %v:%v ", config.Address, config.Port)
	gifBot.Run()
}

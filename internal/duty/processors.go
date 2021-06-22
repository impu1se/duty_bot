package duty

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	tgbotapi "github.com/Syfaro/telegram-bot-api"
	"github.com/impu1se/duty_bot/configs"
)

type System interface {
	ClearDir(pattern string) error
	ReadDutyCSV(file string) ([][]string, error)
}

type GifBot struct {
	Config                *configs.Config
	Updates               tgbotapi.UpdatesChannel
	DutyMonth             [][]string
	system                System
	logger                *zap.Logger
	ctx                   context.Context
	api                   tgbotapi.BotAPI
	mux                   sync.Mutex
	LastNotification      time.Time
	ChatIDForNotification int64
}

func NewDutyBot(
	config *configs.Config,
	updates tgbotapi.UpdatesChannel,
	system System,
	logger *zap.Logger,
	api tgbotapi.BotAPI,
	ctx context.Context,
) *GifBot {
	return &GifBot{
		Config:    config,
		Updates:   updates,
		system:    system,
		logger:    logger,
		ctx:       ctx,
		api:       api,
		DutyMonth: [][]string{},
		mux:       sync.Mutex{},
	}
}

func (bot *GifBot) Run() {
	ticker := time.NewTicker(time.Duration(bot.Config.UpdateTimeInMinutes) * time.Minute)
	bot.logger.Info("Try to read csv on start app")
	bot.readCSV(nil)
	for {
		select {
		case update := <-bot.Updates:
			if update.Message == nil {
				continue
			}
			if update.Message.IsCommand() {
				bot.handlerCommands(&update)
				continue
			}
			if update.Message != nil {
				bot.handlerMessages(&update)
				continue
			}
		case tick := <-ticker.C:
			bot.printSchedule(tick)
		}
	}
}

func (bot *GifBot) NewMessage(chatId int64, message string, button *tgbotapi.ReplyKeyboardMarkup) error {

	if message == "" {
		return nil
	}

	msg := tgbotapi.NewMessage(chatId, message)
	if button != nil {
		msg.ReplyMarkup = button
	}
	if _, err := bot.api.Send(msg); err != nil {
		return err
	}
	return nil
}

func (bot *GifBot) printSchedule(tick time.Time) {
	bot.mux.Lock()
	defer bot.mux.Unlock()
	//baseTime, err := time.Parse("15:04:05", bot.Config.BaseTimeForNotification)
	//if err != nil {
	//	bot.logger.Error(fmt.Sprintf("time is not parsed, err: %v", err.Error()))
	//	return
	//}
	bot.logger.Info(fmt.Sprintf("tick: %v", tick))
	if tick.Hour() == bot.Config.BaseTimeForNotification {
		bot.logger.Info(fmt.Sprintf("tick: %v == base time: %v", tick, bot.Config.BaseTimeForNotification))
		bot.dutyNow(bot.ChatIDForNotification)
		bot.logger.Info(fmt.Sprintf("send notification to chat %v", bot.ChatIDForNotification))
	}

}

func (bot *GifBot) start(update *tgbotapi.Update) {
	bot.mux.Lock()
	bot.ChatIDForNotification = update.Message.Chat.ID
	bot.mux.Unlock()

	err := bot.NewMessage(bot.ChatIDForNotification, "Всем привет я бот для уведомлений дежурных", nil)
	if err != nil {
		bot.logger.Error(fmt.Sprintf("can't send message, err: %v", err))
	}
}

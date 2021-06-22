package duty

import (
	"fmt"
	"time"

	tgbotapi "github.com/Syfaro/telegram-bot-api"
)

const (
	updateDuty = "updateDuty"
	duty       = "duty"
	dutyWeek   = "dutyWeek"
	dutyMonth  = "dutyMonth"
	start      = "start"
)

func (bot *GifBot) handlerCommands(update *tgbotapi.Update) {
	switch update.Message.Command() {
	case start:
		bot.start(update)
	case updateDuty:
		bot.readCSV(update)
	case duty:
		bot.dutyNow(update.Message.Chat.ID)
	case dutyWeek:
	case dutyMonth:
	}
}

func (bot *GifBot) handlerMessages(update *tgbotapi.Update) {
	switch update.Message.Text {
	default:
	}
}

func (bot *GifBot) readCSV(update *tgbotapi.Update) {
	dutyList, err := bot.system.ReadDutyCSV("duty.csv") // TODO: file name
	if err != nil {
		bot.logger.Error(fmt.Sprintf("cant read csv, err: %v", err.Error()))
		return
	}
	bot.mux.Lock()
	bot.DutyMonth = dutyList
	bot.mux.Unlock()
	if update != nil {
		chatId := update.Message.Chat.ID
		if err := bot.NewMessage(chatId, fmt.Sprintf("command %s is done", update.Message.Command()), nil); err != nil {
			bot.logger.Error(fmt.Sprintf("can't send message, reason: %v", err))
		}
	}

	return
}

func (bot *GifBot) dutyNow(chatID int64) {
	timeNow := time.Now()
	bot.mux.Lock()
	defer bot.mux.Unlock()
	for _, duty := range bot.DutyMonth {
		timeDuty, err := time.Parse("2006-01-02", duty[0])
		if err != nil {
			bot.logger.Error(fmt.Sprintf("cant parse time from duty list, err: %v", err.Error()))
			return
		}
		if timeNow.Day() == timeDuty.Day() {
			err := bot.NewMessage(chatID, fmt.Sprintf("сегодня дежурный %s", duty[1]), nil)
			if err != nil {
				bot.logger.Error(err.Error())
				return
			}
		}
	}
}

func checkValidTimes(endTime, startTime int) (string, bool) {
	if endTime <= startTime {
		return "end more start", false
	} else if endTime-startTime > 10 {
		return "video more 10s", false
	}
	return "", true
}

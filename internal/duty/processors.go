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
	ReadDutyCSV(file string) (map[string][][]string, error)
}

type GifBot struct {
	Config                *configs.Config
	Updates               tgbotapi.UpdatesChannel
	DutyMonth             map[string][][]string
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
		Config:                config,
		Updates:               updates,
		system:                system,
		logger:                logger,
		ctx:                   ctx,
		api:                   api,
		DutyMonth:             map[string][][]string{},
		mux:                   sync.Mutex{},
		ChatIDForNotification: -1001428494986, // Chatid == Rnd Alerts
	}
}

func (bot *GifBot) Run() {
	ticker := time.NewTicker(time.Duration(bot.Config.IntervalTime) * time.Minute)
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
	msg.ParseMode = "HTML"
	if _, err := bot.api.Send(msg); err != nil {
		return err
	}
	return nil
}

func (bot *GifBot) printSchedule(tick time.Time) {
	if tick.Hour() == bot.Config.BaseTimeForNotification {
		bot.logger.Info(fmt.Sprintf("tick: %v == base time: %v", tick, bot.Config.BaseTimeForNotification))
		bot.dutyNow(bot.ChatIDForNotification, false)
	}
}

func (bot *GifBot) start(update *tgbotapi.Update) {

	// This for many channel of duty, need to improve
	//bot.mux.Lock()
	//bot.ChatIDForNotification = update.Message.Chat.ID
	//bot.mux.Unlock()

	err := bot.NewMessage(update.Message.Chat.ID, "Всем привет я бот для уведомлений дежурных", nil)
	if err != nil {
		bot.logger.Error(fmt.Sprintf("can't send message, err: %v", err))
	}
}

func (bot *GifBot) getLastDayOfMonth(indexOfDuty, numberOfDay int) int {
	if len(bot.DutyMonth["main"]) < len(bot.DutyMonth["main"])+(indexOfDuty-numberOfDay+8) {
		return len(bot.DutyMonth["main"])
	}
	return indexOfDuty - numberOfDay + 8
}

func (bot *GifBot) getDocument(id int64) {
	err := bot.NewMessage(id, "Документ для дежурного: https://confluence.mail.ru/pages/viewpage.action?pageId=618790765", nil)
	if err != nil {
		bot.logger.Error(fmt.Sprintf("can't send message, err: %v", err))
	}
}

func (bot *GifBot) dutyNextWeek(chatID int64) {
	timeNow := time.Now()
	indexOfDuty := bot.getDutyIndex(timeNow)
	if indexOfDuty == -1 {
		bot.logger.Warn(fmt.Sprintf("cant find duty of day from duty month list"))
		return
	}

	numberOfDay := mappingWeekdays[timeNow.Weekday().String()]
	listOfNextWeekDuty := bot.DutyMonth["main"][indexOfDuty-numberOfDay+8 : indexOfDuty-numberOfDay+8+7]

	var messageWeekDuty string
	for i, weekday := range DaysOfTheWeek {
		messageWeekDuty = messageWeekDuty + generateMessage(i, weekday, listOfNextWeekDuty)
	}

	err := bot.NewMessage(chatID, messageWeekDuty, nil)
	if err != nil {
		bot.logger.Error(err.Error())
	}
}

func (bot *GifBot) dutyWeek(chatID int64, fromCommand bool) {
	timeNow := time.Now()
	indexOfDuty := bot.getDutyIndex(timeNow)
	if indexOfDuty == -1 {
		bot.logger.Warn(fmt.Sprintf("cant find duty of day from duty month list"))
		return
	}

	weekdayNow := timeNow.Weekday().String()
	numberOfDay := mappingWeekdays[weekdayNow]
	lastDayOfMonth := bot.getLastDayOfMonth(indexOfDuty, numberOfDay)
	listOfWeekDutys := bot.DutyMonth["main"][indexOfDuty-numberOfDay+1 : lastDayOfMonth]

	//if len(listOfWeekDutys) != 7 { // TODO: make not fully week of duty list
	//
	//	bot.logger.Info(fmt.Sprintf("len of list of week dutys %v", len(listOfWeekDutys)))
	//	bot.logger.Info(fmt.Sprintf("%v", listOfWeekDutys))
	//	return
	//}

	var messageWeekDuty string
	for i, weekday := range DaysOfTheWeek {
		messageWeekDuty = messageWeekDuty + generateMessage(i, weekday, listOfWeekDutys)
	}

	err := bot.NewMessage(chatID, messageWeekDuty, nil)
	if err != nil {
		bot.logger.Error(err.Error())
	}
}

func (bot *GifBot) getDutyIndex(timeNow time.Time) int {
	indexOfDuty := -1
	for i, duty := range bot.DutyMonth["main"] {
		if len(duty) == 0 {
			continue
		}
		timeDuty, err := time.Parse("2006-01-02", duty[0])
		if err != nil {
			bot.logger.Error(fmt.Sprintf("cant parse time from duty list, err: %v", err.Error()))
			return -1
		}
		if timeNow.Day() == timeDuty.Day() && timeNow.Month() == timeDuty.Month() {
			indexOfDuty = i
			break
		}
	}
	return indexOfDuty
}

func generateMessage(iteration int, weekday string, listOfWeekDuty [][]string) string {
	return fmt.Sprintf("(%v) %v %v %v \n",
		listOfWeekDuty[iteration][0], weekday, listOfWeekDuty[iteration][1], mappingNickName[listOfWeekDuty[iteration][1]])
}

package duty

import (
	"fmt"
	"sort"
	"time"

	tgbotapi "github.com/Syfaro/telegram-bot-api"
)

const (
	updateDuty   = "updateduty"
	duty         = "dutynow"
	dutyWeek     = "dutyweekmain"
	dutyNextWeek = "dutynextweek"
	start        = "start"
	document     = "docs"
	help         = "help"
)

var mappingWeekdays = map[string]int{
	"Monday":    1,
	"Tuesday":   2,
	"Wednesday": 3,
	"Thursday":  4,
	"Friday":    5,
	"Saturday":  6,
	"Sunday":    7,
}

var mappingNickName = map[string]string{
	"aa.grishin":     "@sanagrishin",
	"a.schukin":      "@antonshchukin",
	"a.zavgorodny":   "@aplodismerti",
	"s.yanikin":      "@Impu1s3",
	"i.shvyryalkin":  "@fanyShu",
	"a.novikov":      "@larsnovikov",
	"s.gorustovich":  "@Sulph",
	"pavel.petukhov": "@pashapetukhov",
	"d.chernyshov":   "@dmchr",
	"m.domnin":       "@mark_domnin",
	"d.godov":        "@dgodov",
	"m.ivanova":      "@iv_mari",
	"d.gorev":        "@A2F26",
	"v.kostenko":     "@v_kostenko",
	"g.permyakov":    "@gpermyakov",
	"m.migachev":     "@mMigachev",
	"s.sukhoverhov":  "@ss_now",
	"e.shestakov":    "@p_muzzzko",
}

var DaysOfTheWeek = []string{
	"<b>Понедельник</b>: ",
	"<b>Вторник</b>: ",
	"<b>Среда</b>: ",
	"<b>Четверг</b>: ",
	"<b>Пятница</b>: ",
	"<b>Суббота</b>: ",
	"<b>Воскресенье</b>: "}

var messageOfHelp = `dutynow - Список дежурных на сегодня по командам 
dutyweekmain - Список дежурных по главному направлению на неделю
dutynextweek - Список дежурных по главному направлению на следующую неделю
docs - Справка по алертам`

func (bot *GifBot) handlerCommands(update *tgbotapi.Update) {
	switch update.Message.Command() {
	case start:
		bot.start(update)
	case updateDuty:
		bot.readCSV(update)
	case duty:
		bot.dutyNow(update.Message.Chat.ID, true)
	case dutyWeek:
		bot.dutyWeek(update.Message.Chat.ID, true)
	case dutyNextWeek:
		bot.dutyNextWeek(update.Message.Chat.ID)
	case document:
		bot.getDocument(update.Message.Chat.ID)
	case help:
		bot.NewMessage(update.Message.Chat.ID, messageOfHelp, nil)
	default:
		bot.NewMessage(update.Message.Chat.ID, "Сори не знаю эту команду, посмотри команды с помощью /help", nil)
	}

}

func (bot *GifBot) handlerMessages(update *tgbotapi.Update) {
	//chatId := update.Message.Chat.ID
	//if err := bot.NewMessage(chatId, "hello", nil); err != nil {
	//	bot.logger.Error(fmt.Sprintf("can't send message, reason: %v", err))
	//}
}

func (bot *GifBot) readCSV(update *tgbotapi.Update) {
	dutyList, err := bot.system.ReadDutyCSV("duty_list/duty.csv") // TODO: file name
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

func (bot *GifBot) dutyNow(chatID int64, fromCommand bool) {
	timeNow := time.Now()
	var msgCommand string

	keys := make([]string, 0)
	for key := range bot.DutyMonth {
		if key == "main" {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for i, duty := range bot.DutyMonth["main"] {
		if len(duty) == 0 {
			continue
		}
		timeDuty, err := time.Parse("2006-01-02", duty[0])
		if err != nil {
			bot.logger.Error(fmt.Sprintf("cant parse time from duty list, err: %v", err.Error()))
			return
		}

		if (timeNow.Day() == timeDuty.Day() && timeNow.Month() == timeDuty.Month()) && (bot.LastNotification.Hour() != timeNow.Hour() || fromCommand) {
			msgCommand = msgCommand + fmt.Sprintf(
				"<b>%s</b>: %v %v \n\n", "По направлению", bot.DutyMonth["main"][i][1], mappingNickName[bot.DutyMonth["main"][i][1]])

			for _, commandName := range keys {
				msgCommand = msgCommand + fmt.Sprintf(
					"<b>%s</b>: %v %v \n", commandName, bot.DutyMonth[commandName][i][1], mappingNickName[bot.DutyMonth[commandName][i][1]])
			}
		}

	}
	if msgCommand == "" {
		return
	}

	messageOfDuty := fmt.Sprintf("Сегодня дежурные: \n" + msgCommand)
	err := bot.NewMessage(chatID, messageOfDuty, nil)
	if err != nil {
		bot.logger.Error(err.Error())
	}
	if !fromCommand {
		bot.LastNotification = timeNow
	}
	if timeNow.Weekday().String() == "Monday" && !fromCommand {
		err := bot.NewMessage(chatID, "Так же список дежурных по направлению на неделю:\n", nil)
		if err != nil {
			bot.logger.Error(err.Error())
		}
		bot.dutyWeek(bot.ChatIDForNotification, false)
	}

}

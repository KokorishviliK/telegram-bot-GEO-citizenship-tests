package telegram

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/KokorishviliK/telegram-bot-GEO-citizenship-tests/internal/storage"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
	HelpCmd    = "/help"
	StartCmd   = "/start"
	LngCmd     = "/language"
	HistCmd    = "/history"
	LawBasCmd  = "/lawbasics"
	showAnsCmd = "/showanswers"
)

func (p *Processor) doCmd(text string, chatID int64, command string, messageID int) error {

	text = strings.TrimSpace(text)

	switch command {
	case LngCmd:
		return p.sendTests(storage.GeorgianLanguage, chatID, messageID, text)
	case HistCmd:
		return p.sendTests(storage.History, chatID, messageID, text)
	case LawBasCmd:
		return p.sendTests(storage.LawBasics, chatID, messageID, text)
	case showAnsCmd:
		return p.showAnswers(chatID, messageID, text)
	case HelpCmd:
		return p.sendHelp(chatID)
	case StartCmd:
		return p.sendHello(chatID)
	default:
		return p.sendUnknownCommand(chatID)
	}
}

func (p *Processor) sendTests(topic storage.Topic, chatId int64, messageID int, text string) error {

	questionNum := 1

	topicsComands := topicsComands()
	content := strings.TrimSpace(strings.ReplaceAll(text, topicsComands[topic], ""))

	if content != "" {
		var err error

		questionNum, err = strconv.Atoi(content)
		if err != nil {
			return err
		}
	}

	question, err := p.storage.PickQuestion(topic, questionNum)
	if err != nil {
		return err
	}

	if messageID == 0 {

		msg := tgbotapi.NewMessage(chatId, textQuestion(question))
		msg.ReplyMarkup = p.keyboardMessages(question, false)

		p.bot.Send(msg)

	} else {

		msg := tgbotapi.NewEditMessageText(chatId, messageID, textQuestion(question))
		p.bot.Send(msg)

		msgkeyboard := tgbotapi.NewEditMessageReplyMarkup(chatId, messageID,
			p.keyboardMessages(question, false))

		p.bot.Send(msgkeyboard)
	}

	return nil

}

func (p *Processor) showAnswers(chatId int64, messageID int, text string) error {

	content := strings.TrimSpace(strings.ReplaceAll(text, showAnsCmd, ""))
	partsContent := strings.Split(content, ",")

	topicCode, err := strconv.Atoi(partsContent[0][6:len(partsContent[0])])
	if err != nil {
		return err
	}

	questionNum, err := strconv.Atoi(partsContent[1][9:len(partsContent[1])])
	if err != nil {
		return err
	}

	question, err := p.storage.PickQuestion(storage.Topic(topicCode), questionNum)
	if err != nil {
		return err
	}

	msgkeyboard := tgbotapi.NewEditMessageReplyMarkup(chatId, messageID,
		p.keyboardMessages(question, true))

	p.bot.Send(msgkeyboard)

	return nil

}

func (p *Processor) sendHelp(chatID int64) error {

	msg := tgbotapi.NewMessage(chatID, msgHelp)
	p.bot.Send(msg)

	return nil
}

func (p *Processor) sendUnknownCommand(chatID int64) error {

	msg := tgbotapi.NewMessage(chatID, msgUnknownCommand)
	p.bot.Send(msg)

	return nil
}

func (p *Processor) sendHello(chatID int64) error {

	msg := tgbotapi.NewMessage(chatID, msgHello)
	p.bot.Send(msg)

	return nil

}

func topicsComands() map[storage.Topic]string {

	topicsComands := map[storage.Topic]string{
		storage.GeorgianLanguage: LngCmd,
		storage.History:          HistCmd,
		storage.LawBasics:        LawBasCmd,
	}

	return topicsComands

}

func textQuestion(question *storage.Question) string {

	textQuestion := ""

	buffer := bytes.Buffer{}
	buffer.WriteString(question.Text)

	for _, answer := range question.Answers {
		text := "\n\n" + answer.Code + ". " + answer.Text
		buffer.WriteString(text)
	}

	textQuestion = buffer.String()

	return textQuestion
}

func (p *Processor) keyboardMessages(question *storage.Question, showAnswers bool) tgbotapi.InlineKeyboardMarkup {

	keyboardMessages := &tgbotapi.InlineKeyboardMarkup{}

	thisLastQuestion := p.storage.QuantityQuestions(question.Topic) == question.Number

	addAnswersToKeyboard(keyboardMessages, question, showAnswers, thisLastQuestion)

	addСontrolButtonsToKeyboard(keyboardMessages, question, showAnswers, thisLastQuestion)

	return *keyboardMessages

}

func addAnswersToKeyboard(keyboardMessages *tgbotapi.InlineKeyboardMarkup, question *storage.Question,
	showAnswers bool, thisLastQuestion bool) {

	dataBtnShowAns := dataBtnShowAns(question)
	dataBtnNextQuestion := dataBtnNextQuestion(question, thisLastQuestion)

	var rowAnswers []tgbotapi.InlineKeyboardButton

	for _, answer := range question.Answers {

		dataBtn := ""
		btnText := answer.Code

		if answer.ThisCorrectAnswer {
			dataBtn = dataBtnNextQuestion
			if showAnswers {

				btnText = btnText + " ✅"
			}
		} else {
			dataBtn = dataBtnShowAns
		}

		btn := tgbotapi.NewInlineKeyboardButtonData(btnText, dataBtn)
		rowAnswers = append(rowAnswers, btn)
	}

	keyboardMessages.InlineKeyboard = append(keyboardMessages.InlineKeyboard, rowAnswers)

}

func addСontrolButtonsToKeyboard(keyboardMessages *tgbotapi.InlineKeyboardMarkup, question *storage.Question,
	showAnswers bool, thisLastQuestion bool) {

	topicsComands := topicsComands()
	topicComand := topicsComands[question.Topic]

	dataBtnShowAns := dataBtnShowAns(question)
	dataBtnNextQuestion := dataBtnNextQuestion(question, thisLastQuestion)

	var row []tgbotapi.InlineKeyboardButton

	if question.Number > 1 {
		btnPrevious := tgbotapi.NewInlineKeyboardButtonData("<<",
			topicComand+" "+strconv.Itoa(question.Number-1))

		row = append(row, btnPrevious)
	}

	btnShowAns := tgbotapi.NewInlineKeyboardButtonData("show answers", dataBtnShowAns)
	row = append(row, btnShowAns)

	if !thisLastQuestion {
		btnNext := tgbotapi.NewInlineKeyboardButtonData(">>", dataBtnNextQuestion)
		row = append(row, btnNext)
	}

	keyboardMessages.InlineKeyboard = append(keyboardMessages.InlineKeyboard, row)

}

func dataBtnShowAns(question *storage.Question) string {

	return fmt.Sprintf("%v topic:%v,question:%v",
		showAnsCmd, question.Topic, strconv.Itoa(question.Number))

}

func dataBtnNextQuestion(question *storage.Question, thisLastQuestion bool) string {

	var nextQuestion int

	if thisLastQuestion {
		nextQuestion = 1
	} else {
		nextQuestion = question.Number + 1
	}

	topicsComands := topicsComands()
	topicComand := topicsComands[question.Topic]

	return topicComand + " " + strconv.Itoa(nextQuestion)

}

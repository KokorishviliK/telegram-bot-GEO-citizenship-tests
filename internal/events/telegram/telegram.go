package telegram

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/KokorishviliK/telegram-bot-GEO-citizenship-tests/internal/events"
	"github.com/KokorishviliK/telegram-bot-GEO-citizenship-tests/internal/storage"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type Processor struct {
	bot     *tgbotapi.BotAPI
	offset  int
	storage storage.Storage
}

type Meta struct {
	ChatID       int64
	MessageID    int
	Command      string
	CallbackData string
}

var (
	ErrUnknownEventType = errors.New("unknown event type")
	ErrUnknownMetaType  = errors.New("unknown meta type")
)

func New(client *tgbotapi.BotAPI, storage storage.Storage) *Processor {
	return &Processor{
		bot:     client,
		storage: storage,
	}
}

func (p *Processor) Fetch(limit int) ([]events.Event, error) {

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 1
	u.Limit = limit
	u.Offset = p.offset

	updates, err := p.bot.GetUpdates(u)
	if err != nil {
		log.Panic(err)
	}

	if len(updates) == 0 {
		return nil, nil
	}

	res := make([]events.Event, 0, len(updates))

	for _, update := range updates {

		res = append(res, event(update))

	}

	p.offset = updates[len(updates)-1].UpdateID + 1

	return res, nil
}

func (p *Processor) HandleUpdate(r *http.Request) error {

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	var update tgbotapi.Update
	err = json.Unmarshal(body, &update)
	if err != nil {
		return err
	}

	event := event(update)
	err = p.Process(event)

	return err
}

func (p *Processor) Process(event events.Event) error {

	switch event.Type {
	case events.Message:
		return p.processMessage(event)
	case events.Callback:
		return p.processCallback(event)
	default:
		return fmt.Errorf("can't process event %w", ErrUnknownEventType)
	}
}

func (p *Processor) processMessage(event events.Event) error {

	meta, err := meta(event)
	if err != nil {
		return fmt.Errorf("can't process message %w", err)
	}

	command := "/" + meta.Command

	if err := p.doCmd(event.Text, meta.ChatID, command, 0); err != nil {
		return fmt.Errorf("can't process message %w", err)
	}

	return err
}

func (p *Processor) processCallback(event events.Event) error {

	meta, err := meta(event)
	if err != nil {
		return fmt.Errorf("can't process сallback %w", err)
	}

	command := callbackDataCommand(meta.CallbackData)

	if err := p.doCmd(meta.CallbackData, meta.ChatID, command, meta.MessageID); err != nil {
		return fmt.Errorf("can't process сallback %w", err)
	}

	return err
}

func event(u tgbotapi.Update) events.Event {

	eventType := eventType(u)

	event := events.Event{
		Type: eventType,
		Text: fetchText(u),
	}

	var message *tgbotapi.Message
	сallbackData := ""

	switch event.Type {
	case events.Message:
		message = u.Message
	case events.Callback:
		message = u.CallbackQuery.Message
		сallbackData = u.CallbackQuery.Data
	}

	event.Meta = Meta{
		ChatID:       message.Chat.ID,
		MessageID:    message.MessageID,
		Command:      message.Command(),
		CallbackData: сallbackData,
	}

	return event
}

func callbackDataCommand(callbackData string) string {

	var callbackDataCommand string

	if strings.HasPrefix(callbackData, LngCmd) {
		callbackDataCommand = LngCmd
	} else if strings.HasPrefix(callbackData, HistCmd) {
		callbackDataCommand = HistCmd
	} else if strings.HasPrefix(callbackData, LawBasCmd) {
		callbackDataCommand = LawBasCmd
	} else if strings.HasPrefix(callbackData, showAnsCmd) {
		callbackDataCommand = showAnsCmd
	}

	return callbackDataCommand
}

func meta(event events.Event) (Meta, error) {

	meta, ok := event.Meta.(Meta)
	if !ok {
		return Meta{}, fmt.Errorf("can't get meta %w", ErrUnknownMetaType)
	}

	return meta, nil
}

func fetchText(upd tgbotapi.Update) string {

	if upd.Message == nil {
		return ""
	}

	return upd.Message.Text
}

func eventType(upd tgbotapi.Update) events.Type {

	eventType := events.Unknown

	if upd.Message != nil {
		eventType = events.Message
	} else if upd.CallbackQuery != nil {
		eventType = events.Callback
	}

	return eventType
}

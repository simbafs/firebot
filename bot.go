package main

import (
	"log"
	"strings"
	"time"

	"tainanfire/bucket"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

type Msg struct {
	event *Event
	msg   *gotgbot.Message
}

type Bot struct {
	bot    *gotgbot.Bot
	bucket *bucket.Bucket[string, Msg]
}

type BotOpt struct {
	APIKey    string
	AliveTime time.Duration
}

func WithAPIKey(apikey string) func(*BotOpt) {
	return func(opt *BotOpt) {
		opt.APIKey = apikey
	}
}

func WithAliveTime(d time.Duration) func(*BotOpt) {
	return func(opt *BotOpt) {
		opt.AliveTime = d
	}
}

func NewBot(opts ...func(*BotOpt)) *Bot {
	defaultOpt := &BotOpt{
		AliveTime: 48 * time.Hour,
	}

	for _, opt := range opts {
		opt(defaultOpt)
	}

	bot, err := gotgbot.NewBot(defaultOpt.APIKey, nil)
	if err != nil {
		panic(err)
	}
	return &Bot{
		bot:    bot,
		bucket: bucket.New[string, Msg](defaultOpt.AliveTime),
	}
}

func (b *Bot) SendMessage(chat int64, msg string) (*gotgbot.Message, error) {
	// escape markdown
	msg = strings.ReplaceAll(msg, "-", "\\-")
	return b.bot.SendMessage(chat, msg, &gotgbot.SendMessageOpts{
		ParseMode: "MarkdownV2",
	})
}

func (b *Bot) SendEvent(chat int64, e *Event, first bool) error {
	if first {
		b.bucket.Set(e.Key, Msg{
			event: e,
			msg:   nil,
		})
		return nil
	}
	oldMsg, ok := b.bucket.Get(e.Key)
	var text string
	var msg *gotgbot.Message
	var err error

	if !ok {
		// New event
		text += "新事件\n" + e.String()
		msg, err = b.SendMessage(chat, text)
	} else if diff := oldMsg.event.Diff(e); diff != "" {
		// update old event
		text += "事件更新\n" + diff + "\n" + e.String()
		msg, err = b.SendMessage(chat, text)
	} else {
		return nil
	}

	log.Println(text)

	if err != nil {
		return err
	}

	b.bucket.Set(e.Key, Msg{
		event: e,
		msg:   msg,
	})

	return nil
}

func (b *Bot) GC() {
	b.bucket.GC()
}

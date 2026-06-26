package main

import (
	"fmt"
	"log"
	"sync"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

type Bot interface {
	Broadcast(chat int64, result DiffResult, silent bool) error
	UnpinAll(chat int64) error
}

type TGBot struct {
	bot       *gotgbot.Bot
	mu        sync.Mutex
	msgIDs    map[string]int64      // UID → pinned Telegram message_id
	eventRows map[string][]eventRow // UID → accumulated table rows
}

type BotOpt struct {
	APIKey string
}

func WithAPIKey(apikey string) func(*BotOpt) {
	return func(opt *BotOpt) {
		opt.APIKey = apikey
	}
}

func NewTGBot(opts ...func(*BotOpt)) *TGBot {
	defaultOpt := &BotOpt{}

	for _, opt := range opts {
		opt(defaultOpt)
	}

	bot, err := gotgbot.NewBot(defaultOpt.APIKey, nil)
	if err != nil {
		panic(err)
	}
	return &TGBot{
		bot:       bot,
		msgIDs:    make(map[string]int64),
		eventRows: make(map[string][]eventRow),
	}
}

func (b *TGBot) storeMsgID(uid string, msgID int64) {
	b.mu.Lock()
	b.msgIDs[uid] = msgID
	b.mu.Unlock()
}

func (b *TGBot) getMsgID(uid string) (int64, bool) {
	b.mu.Lock()
	id, ok := b.msgIDs[uid]
	b.mu.Unlock()
	return id, ok
}

func (b *TGBot) delMsgID(uid string) {
	b.mu.Lock()
	delete(b.msgIDs, uid)
	b.mu.Unlock()
}

func (b *TGBot) storeRows(uid string, rows []eventRow) {
	b.mu.Lock()
	b.eventRows[uid] = rows
	b.mu.Unlock()
}

func (b *TGBot) getRows(uid string) ([]eventRow, bool) {
	b.mu.Lock()
	rows, ok := b.eventRows[uid]
	b.mu.Unlock()
	return rows, ok
}

func (b *TGBot) delRows(uid string) {
	b.mu.Lock()
	delete(b.eventRows, uid)
	b.mu.Unlock()
}

func (b *TGBot) UnpinAll(chat int64) error {
	_, err := b.bot.UnpinAllChatMessages(chat, nil)
	return err
}

func (b *TGBot) Broadcast(chat int64, result DiffResult, silent bool) error {
	for i := range result.New {
		event := &result.New[i]
		h := heading(event.Location, event.Category)
		rows := []eventRow{event.initialRow()}
		markdown := renderRows(h, "🆕 新事件", rows)

		msg, err := sendRichMessage(b.bot.Token, chat, markdown, silent)
		if err != nil {
			return err
		}

		if _, err := b.bot.PinChatMessage(chat, msg.MessageId, &gotgbot.PinChatMessageOpts{
			DisableNotification: true,
		}); err != nil {
			log.Println("[pin-err]", event.UID, err)
		}

		b.storeMsgID(event.UID, msg.MessageId)
		b.storeRows(event.UID, rows)
		log.Println("[new]", event.UID)
	}

	for i := range result.Updated {
		ed := &result.Updated[i]
		h := heading(ed.New.Location, ed.New.Category)
		activity := activityLine(ed.Changes)

		prevRows, _ := b.getRows(ed.New.UID)
		newRow := snapshotRow(ed.New.Status, ed.New.Brigade.String())
		rows := append(prevRows, newRow)
		markdown := renderRows(h, activity, rows)

		msgID, ok := b.getMsgID(ed.New.UID)
		if ok {
			if _, err := editRichMessage(b.bot.Token, chat, msgID, markdown); err != nil {
				log.Println("[edit-err]", ed.New.UID, err)
				// Fallback: send as new message + pin (always notify on update)
				msg, err2 := sendRichMessage(b.bot.Token, chat, markdown, false)
				if err2 != nil {
					return err2
				}
				if _, err := b.bot.PinChatMessage(chat, msg.MessageId, &gotgbot.PinChatMessageOpts{
					DisableNotification: true,
				}); err != nil {
					log.Println("[pin-err]", ed.New.UID, err)
				}
				b.storeMsgID(ed.New.UID, msg.MessageId)
			}
		} else {
			// Previously untracked — treat as new
			msg, err := sendRichMessage(b.bot.Token, chat, markdown, false)
			if err != nil {
				return err
			}
			if _, err := b.bot.PinChatMessage(chat, msg.MessageId, &gotgbot.PinChatMessageOpts{
				DisableNotification: true,
			}); err != nil {
				log.Println("[pin-err]", ed.New.UID, err)
			}
			b.storeMsgID(ed.New.UID, msg.MessageId)
		}
		b.storeRows(ed.New.UID, rows)
		log.Println("[update]", ed.New.UID)
	}

	for i := range result.Deleted {
		event := &result.Deleted[i]
		msgID, ok := b.getMsgID(event.UID)
		if !ok {
			continue
		}

		prevRows, _ := b.getRows(event.UID)
		finalRow := snapshotRow("已結案", event.Brigade.String())
		rows := append(prevRows, finalRow)
		h := heading(event.Location, event.Category)
		markdown := renderRows(h, "已結案", rows)

		if _, err := editRichMessage(b.bot.Token, chat, msgID, markdown); err != nil {
			log.Println("[close-edit-err]", event.UID, err)
		}

		if _, err := b.bot.UnpinChatMessage(chat, &gotgbot.UnpinChatMessageOpts{
			MessageId: &msgID,
		}); err != nil {
			log.Println("[unpin-err]", event.UID, err)
		}
		b.delMsgID(event.UID)
		b.delRows(event.UID)
		log.Println("[close]", event.UID)
	}

	return nil
}

type LocalBot struct{}

func NewLocalBot() *LocalBot {
	return &LocalBot{}
}

func (b *LocalBot) UnpinAll(chat int64) error {
	fmt.Printf("[Chat %d] unpin all\n", chat)
	return nil
}

func (b *LocalBot) Broadcast(chat int64, result DiffResult, silent bool) error {
	for i := range result.New {
		event := &result.New[i]
		h := heading(event.Location, event.Category)
		rows := []eventRow{event.initialRow()}
		fmt.Printf("[Chat %d] [new] %s\n%s\n\n", chat, event.UID, renderRows(h, "🆕 新事件", rows))
	}
	for i := range result.Updated {
		ed := &result.Updated[i]
		h := heading(ed.New.Location, ed.New.Category)
		activity := activityLine(ed.Changes)
		row := snapshotRow(ed.New.Status, ed.New.Brigade.String())
		fmt.Printf("[Chat %d] [update] %s\n%s\n\n", chat, ed.New.UID, renderRows(h, activity, []eventRow{row}))
	}
	for i := range result.Deleted {
		event := &result.Deleted[i]
		h := heading(event.Location, event.Category)
		row := snapshotRow("已結案", event.Brigade.String())
		fmt.Printf("[Chat %d] [close] %s\n%s\n\n", chat, event.UID, renderRows(h, "已結案", []eventRow{row}))
	}
	return nil
}

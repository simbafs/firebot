package telegram

import (
	"fmt"
	"log"
	"sync"

	"tainanfire/diff"
	"tainanfire/render"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

type Bot interface {
	Broadcast(chat int64, result diff.DiffResult, silent bool) error
	UnpinAll(chat int64) error
}

type TGBot struct {
	bot       *gotgbot.Bot
	mu        sync.Mutex
	msgIDs    map[string]int64
	eventRows map[string][]render.EventRow
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
		eventRows: make(map[string][]render.EventRow),
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

func (b *TGBot) storeRows(uid string, rows []render.EventRow) {
	b.mu.Lock()
	b.eventRows[uid] = rows
	b.mu.Unlock()
}

func (b *TGBot) getRows(uid string) ([]render.EventRow, bool) {
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

func (b *TGBot) closeEvent(chat int64, uid, location, category, brigade string) {
	msgID, ok := b.getMsgID(uid)
	if !ok {
		return
	}

	prevRows, _ := b.getRows(uid)
	finalRow := render.SnapshotRow("已結案", brigade)
	rows := append(prevRows, finalRow)
	h := render.Heading(location, category)
	markdown := render.RenderRows(h, "已結案", rows)

	if _, err := editRichMessage(b.bot.Token, chat, msgID, markdown); err != nil {
		log.Println("[close-edit-err]", uid, err)
	}

	if _, err := b.bot.UnpinChatMessage(chat, &gotgbot.UnpinChatMessageOpts{
		MessageId: &msgID,
	}); err != nil {
		log.Println("[unpin-err]", uid, err)
	}
	b.delMsgID(uid)
	b.delRows(uid)
}

func (b *TGBot) Broadcast(chat int64, result diff.DiffResult, silent bool) error {
	for i := range result.New {
		event := &result.New[i]
		h := render.Heading(event.Location, event.Category)
		rows := []render.EventRow{render.InitialRow(event)}
		markdown := render.RenderRows(h, "🆕 新事件", rows)

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

		// "火已滅" means the fire is out — treat as closed immediately.
		if ed.New.Status == "火已滅" {
			b.closeEvent(chat, ed.New.UID, ed.New.Location, ed.New.Category, ed.New.Brigade.String())
			log.Println("[close]", ed.New.UID)
			continue
		}

		h := render.Heading(ed.New.Location, ed.New.Category)
		activity := render.ActivityLine(ed.Changes)

		prevRows, _ := b.getRows(ed.New.UID)
		newRow := render.SnapshotRow(ed.New.Status, ed.New.Brigade.String())
		rows := append(prevRows, newRow)
		markdown := render.RenderRows(h, activity, rows)

		msgID, ok := b.getMsgID(ed.New.UID)
		if ok {
			if _, err := editRichMessage(b.bot.Token, chat, msgID, markdown); err != nil {
				log.Println("[edit-err]", ed.New.UID, err)
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
		b.closeEvent(chat, event.UID, event.Location, event.Category, event.Brigade.String())
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

func (b *LocalBot) Broadcast(chat int64, result diff.DiffResult, silent bool) error {
	for i := range result.New {
		event := &result.New[i]
		h := render.Heading(event.Location, event.Category)
		rows := []render.EventRow{render.InitialRow(event)}
		fmt.Printf("[Chat %d] [new] %s\n%s\n\n", chat, event.UID, render.RenderRows(h, "🆕 新事件", rows))
	}
	for i := range result.Updated {
		ed := &result.Updated[i]
		h := render.Heading(ed.New.Location, ed.New.Category)
		activity := render.ActivityLine(ed.Changes)
		row := render.SnapshotRow(ed.New.Status, ed.New.Brigade.String())
		fmt.Printf("[Chat %d] [update] %s\n%s\n\n", chat, ed.New.UID, render.RenderRows(h, activity, []render.EventRow{row}))
	}
	for i := range result.Deleted {
		event := &result.Deleted[i]
		h := render.Heading(event.Location, event.Category)
		row := render.SnapshotRow("已結案", event.Brigade.String())
		fmt.Printf("[Chat %d] [close] %s\n%s\n\n", chat, event.UID, render.RenderRows(h, "已結案", []render.EventRow{row}))
	}
	return nil
}

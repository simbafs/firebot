package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

const apiURL = "https://api.telegram.org"

// sendRichMessage sends a rich formatted message using Bot API 10.1 sendRichMessage.
// gotgbot does not yet support this endpoint, so we call it via raw HTTP.
// mapsURL, if non-empty, adds an inline keyboard button that opens the given URL
// (since & in URLs gets HTML-escaped by rich_message's markdown parser, the
// inline keyboard button's url field bypasses this).
func sendRichMessage(token string, chatID int64, markdown string, mapsURL string, silent bool) (*gotgbot.Message, error) {
	body := map[string]any{
		"chat_id": chatID,
		"rich_message": map[string]string{
			"markdown": markdown,
		},
	}
	if silent {
		body["disable_notification"] = true
	}
	if mapsURL != "" {
		body["reply_markup"] = map[string]any{
			"inline_keyboard": [][]map[string]string{
				{{"text": "📍 Google 地圖", "url": mapsURL}},
			},
		}
	}
	return apiPost(token, "sendRichMessage", body)
}

// editRichMessage edits an existing rich message in-place.
// Uses editMessageText with the rich_message parameter (Bot API 10.1).
// mapsURL, if non-empty, adds/updates the inline keyboard button.
func editRichMessage(token string, chatID int64, msgID int64, markdown string, mapsURL string) (*gotgbot.Message, error) {
	body := map[string]any{
		"chat_id":    chatID,
		"message_id": msgID,
		"rich_message": map[string]string{
			"markdown": markdown,
		},
	}
	if mapsURL != "" {
		body["reply_markup"] = map[string]any{
			"inline_keyboard": [][]map[string]string{
				{{"text": "📍 Google 地圖", "url": mapsURL}},
			},
		}
	}
	return apiPost(token, "editMessageText", body)
}

func apiPost(token string, method string, body any) (*gotgbot.Message, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal %s body: %w", method, err)
	}

	url := fmt.Sprintf("%s/bot%s/%s", apiURL, token, method)

	// Retry transient network errors (connection reset, timeout) up to 3 times
	// with short backoff. Telegram API errors (e.g. 403, message not found) are
	// returned immediately without retry.
	var resp *http.Response
	for attempt := range 3 {
		resp, err = http.Post(url, "application/json", bytes.NewReader(b))
		if err == nil {
			break
		}
		if attempt < 2 {
			time.Sleep(time.Duration(attempt+1) * time.Second)
		}
	}
	if err != nil {
		return nil, fmt.Errorf("%s request: %w", method, err)
	}
	defer resp.Body.Close()

	var r gotgbot.Response
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, fmt.Errorf("%s decode: %w", method, err)
	}

	if !r.Ok {
		return nil, &gotgbot.TelegramError{
			Method:      method,
			Code:        r.ErrorCode,
			Description: r.Description,
		}
	}

	var msg gotgbot.Message
	if err := json.Unmarshal(r.Result, &msg); err != nil {
		return nil, fmt.Errorf("%s unmarshal message: %w", method, err)
	}

	return &msg, nil
}

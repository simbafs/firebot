package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

const apiURL = "https://api.telegram.org"

// sendRichMessage sends a rich formatted message using Bot API 10.1 sendRichMessage.
// gotgbot does not yet support this endpoint, so we call it via raw HTTP.
// When silent is true, the message is sent without notification (used for initial bulk load).
func sendRichMessage(token string, chatID int64, markdown string, silent bool) (*gotgbot.Message, error) {
	body := map[string]any{
		"chat_id": chatID,
		"rich_message": map[string]string{
			"markdown": markdown,
		},
	}
	if silent {
		body["disable_notification"] = true
	}
	return apiPost(token, "sendRichMessage", body)
}

// editRichMessage edits an existing rich message in-place.
// Uses editMessageText with the rich_message parameter (Bot API 10.1).
func editRichMessage(token string, chatID int64, msgID int64, markdown string) (*gotgbot.Message, error) {
	body := map[string]any{
		"chat_id":    chatID,
		"message_id": msgID,
		"rich_message": map[string]string{
			"markdown": markdown,
		},
	}
	return apiPost(token, "editMessageText", body)
}

func apiPost(token string, method string, body any) (*gotgbot.Message, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal %s body: %w", method, err)
	}

	url := fmt.Sprintf("%s/bot%s/%s", apiURL, token, method)
	resp, err := http.Post(url, "application/json", bytes.NewReader(b))
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

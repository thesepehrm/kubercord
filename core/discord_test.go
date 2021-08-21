package core

import "testing"

func TestWebhookToTokens(t *testing.T) {
	wId, wToken := webhookURLTotokens("https://discordapp.com/api/webhooks/12345678901234567/abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890/")
	if wId != 12345678901234567 {
		t.Errorf("Webhook ID is not correct: %d", wId)
	}
	if wToken != "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890" {
		t.Errorf("Webhook token is not correct: %s", wToken)
	}

}

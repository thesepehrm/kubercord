package core

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/andersfylling/snowflake"
	"github.com/nickname32/discordhook"
	"github.com/sirupsen/logrus"
	"github.com/thesepehrm/kubercord/config"
)

type DiscordHandler struct {
	wa *discordhook.WebhookAPI
	wh *discordhook.Webhook
}

func NewDiscordHandler() (*DiscordHandler, error) {

	wId, wToken := webhookURLTotokens(config.Config.Discord.Webhook)

	wa, err := discordhook.NewWebhookAPI(snowflake.Snowflake(wId), wToken, true, nil)
	if err != nil {
		return nil, err
	}
	wh, err := wa.Get(context.Background())
	if err != nil {
		return nil, err
	}

	logrus.WithField("name", wh.Name).Info("Discord webhook initialized")
	return &DiscordHandler{wa, wh}, nil
}

func (dh *DiscordHandler) SendAlert(alert *Alert) error {

	if len(alert.logs) == 0 {
		return nil
	}
	msg := buildMessage(alert)
	if msg == "" {
		return nil
	}
	_, err := dh.wa.Execute(context.Background(), &discordhook.WebhookExecuteParams{
		Embeds: []*discordhook.Embed{
			{
				Title:       alert.service,
				Description: msg,
				Color:       alert.level.Color(),
				Timestamp:   &alert.timestamp,
			},
		},
	}, nil, "")

	return err
}

func (dh *DiscordHandler) SendPodStatus(service, status, reason string) error {
	_, err := dh.wa.Execute(context.Background(), &discordhook.WebhookExecuteParams{
		Embeds: []*discordhook.Embed{
			{
				Title:       service,
				Description: "Pod status changed to: " + status + "\nMessage: " + reason,
				Color:       0xffd700,
			},
		},
	}, nil, "")
	return err
}

func webhookURLTotokens(url string) (int64, string) {
	token := strings.TrimSuffix(url, "/")
	tokenSplit := strings.Split(token, "/")
	return parseInt64(tokenSplit[len(tokenSplit)-2]), tokenSplit[len(tokenSplit)-1]
}

func parseInt64(str string) int64 {
	i, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0
	}
	return i
}

func buildMessage(alert *Alert) string {
	if len(alert.logs) == 0 {
		return ""
	}
	return fmt.Sprintf("Level: %s\nMessage: %s\n\n*Latest logs:*\n```bash\n%s```", alert.level.String(), alert.msg, strings.Join(alert.logs, "\n"))
}

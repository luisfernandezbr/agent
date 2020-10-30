package slack

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/nlopes/slack"
)

// Client _
type Client interface {
	SendMessage(msg string, args ...interface{}) error
}

// New creates a new instance
func New(slackToken string, slackChannel string) (Client, error) {
	slackClient := slack.New(slackToken)
	if _, err := slackClient.AuthTest(); err != nil {
		return nil, err
	}
	var slackChannelID string
	var cursor string
	for {
		var err error
		var channels []slack.Channel
		// needed scopes: channels:read, groups:read, im:read, mpim:read
		channels, cursor, err = slackClient.GetConversations(&slack.GetConversationsParameters{
			ExcludeArchived: "true",
			Limit:           1000,
			Types:           []string{"public_channel", "private_channel"},
			Cursor:          cursor,
		})
		if err != nil {
			return nil, fmt.Errorf("error getting channel ids from slack. error: %v", err)
		}
		if len(channels) == 0 {
			return nil, fmt.Errorf("error getting channel ids from slack")
		}

		for _, c := range channels {
			if c.Name == slackChannel {
				slackChannelID = c.ID
				break
			}
		}
		if slackChannelID != "" {
			break
		}
	}
	if slackChannelID == "" {
		return nil, fmt.Errorf("error finding slack channel %s", slackChannel)
	}
	return &client{
		slackChannelID: slackChannelID,
		slackClient:    slackClient,
	}, nil
}

type client struct {
	slackChannelID string
	slackClient    *slack.Client
}

// SendMessage sends a message to our slack channel "processing_events"
func (c *client) SendMessage(msg string, args ...interface{}) error {
	if c.slackChannelID == "" {
		return nil
	}
	parts := []string{}
	var val string
	for i, m := range args {
		if i%2 != 0 {
			b, _ := json.Marshal(m)
			val += "=" + string(b)
			parts = append(parts, val)
		} else {
			val = fmt.Sprint(m)
		}
	}
	if len(args)%2 != 0 {
		val += "=(MISSING)"
		parts = append(parts, val)
	}
	sort.Strings(parts)
	cnt := strings.Join(parts, "\n")
	// needed scopes: chat:write
	if _, _, err := c.slackClient.PostMessageContext(context.Background(),
		c.slackChannelID,
		slack.MsgOptionText(msg+" ```"+cnt+"```", false),
		slack.MsgOptionAsUser(false),
	); err != nil {
		return err
	}
	return nil
}

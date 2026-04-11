package slack

import (
	"github.com/sipeed/trenchlaw/pkg/bus"
	"github.com/sipeed/trenchlaw/pkg/channels"
	"github.com/sipeed/trenchlaw/pkg/config"
)

func init() {
	channels.RegisterFactory("slack", func(cfg *config.Config, b *bus.MessageBus) (channels.Channel, error) {
		return NewSlackChannel(cfg.Channels.Slack, b)
	})
}

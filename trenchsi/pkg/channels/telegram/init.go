package telegram

import (
	"github.com/sipeed/trenchlaw/pkg/bus"
	"github.com/sipeed/trenchlaw/pkg/channels"
	"github.com/sipeed/trenchlaw/pkg/config"
)

func init() {
	channels.RegisterFactory("telegram", func(cfg *config.Config, b *bus.MessageBus) (channels.Channel, error) {
		return NewTelegramChannel(cfg, b)
	})
}

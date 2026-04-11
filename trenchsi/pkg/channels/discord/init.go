package discord

import (
	"github.com/sipeed/trenchlaw/pkg/bus"
	"github.com/sipeed/trenchlaw/pkg/channels"
	"github.com/sipeed/trenchlaw/pkg/config"
)

func init() {
	channels.RegisterFactory("discord", func(cfg *config.Config, b *bus.MessageBus) (channels.Channel, error) {
		return NewDiscordChannel(cfg.Channels.Discord, b)
	})
}

package jame

import (
	"github.com/sipeed/trenchlaw/pkg/bus"
	"github.com/sipeed/trenchlaw/pkg/channels"
	"github.com/sipeed/trenchlaw/pkg/config"
)

func init() {
	channels.RegisterFactory("jame", func(cfg *config.Config, b *bus.MessageBus) (channels.Channel, error) {
		return NewJameChannel(cfg.Channels.Jame, b)
	})
	channels.RegisterFactory("jame_client", func(cfg *config.Config, b *bus.MessageBus) (channels.Channel, error) {
		return NewJameClientChannel(cfg.Channels.JameClient, b)
	})
}

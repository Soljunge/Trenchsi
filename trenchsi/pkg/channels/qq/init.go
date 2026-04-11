package qq

import (
	"github.com/sipeed/trenchlaw/pkg/bus"
	"github.com/sipeed/trenchlaw/pkg/channels"
	"github.com/sipeed/trenchlaw/pkg/config"
)

func init() {
	channels.RegisterFactory("qq", func(cfg *config.Config, b *bus.MessageBus) (channels.Channel, error) {
		return NewQQChannel(cfg.Channels.QQ, b)
	})
}

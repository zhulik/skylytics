package nats

import (
	"context"
	"log/slog"
	"skylytics/internal/config"
	"time"

	libnats "github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

const (
	appName = "skylytics"
)

type NATS struct {
	Logger *slog.Logger
	Config *config.Config

	JS jetstream.JetStream
	KV jetstream.KeyValue
}

func (n *NATS) Init(ctx context.Context) error {
	nc, err := libnats.Connect(n.Config.NATSURL)
	if err != nil {
		return err
	}

	js, err := jetstream.New(nc)
	if err != nil {
		return err
	}

	n.JS = js

	if n.Config.NATSInit {
		if err := n.initNATS(ctx); err != nil {
			return err
		}
	}

	kv, err := js.KeyValue(ctx, appName)
	if err != nil {
		return err
	}
	n.KV = kv

	return nil
}

func (n *NATS) HealthCheck(context.Context) error {
	_, err := n.JS.Conn().RTT()
	return err
}

func (n *NATS) Shutdown(context.Context) error {
	return n.JS.Conn().Drain()
}

func (n *NATS) initNATS(ctx context.Context) error {
	n.Logger.Info("Initializing NATS")
	_, err := n.JS.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:     appName,
		Subjects: []string{appName + ".*"},
		MaxAge:   24 * time.Hour,
	})
	if err != nil {
		return err
	}
	n.Logger.Info("Stream created or updated", "name", appName)

	_, err = n.JS.CreateOrUpdateKeyValue(ctx, jetstream.KeyValueConfig{
		Bucket: appName,
	})
	if err != nil {
		return err
	}
	n.Logger.Info("KeyValue created or updated", "name", appName)

	return nil
}

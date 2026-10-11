package listener

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/Alessandro-Massarotti-Jr/go-api/internal/config"
	"github.com/Alessandro-Massarotti-Jr/go-api/internal/providers"
)

type Listener struct {
	logger *config.Logger
}

func Create() *Listener {
	return &Listener{logger: config.GetInstance()}
}

// Start registers every listener on the RabbitMQ event provider and consumes events
// until SIGINT/SIGTERM, letting in-flight events finish before returning.
func (listener *Listener) Start() error {
	eventProvider, err := providers.GetRabbitMQEventProviderInstance()
	if err != nil {
		return err
	}

	RegisterListeners(RegisterListenersInput{EventProvider: eventProvider})

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	listener.logger.Info(config.LogMessage{Action: "listener.start", Message: "Listener started"})
	err = eventProvider.StartListening(ctx)
	listener.logger.Info(config.LogMessage{Action: "listener.shutdown", Message: "Listener stopped"})

	return err
}

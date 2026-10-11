package server

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Alessandro-Massarotti-Jr/go-api/internal/config"
	"github.com/Alessandro-Massarotti-Jr/go-api/internal/logger"
	"github.com/Alessandro-Massarotti-Jr/go-api/internal/middlewares"
	"github.com/gin-gonic/gin"
)

type Server struct {
	httpServer *http.Server
	logger     *logger.Logger
}

// Create builds the gin engine with every route registered, listening on the configured PORT.
func Create() *Server {
	engine := gin.New()
	engine.Use(gin.Recovery(), middlewares.ErrorHandler())
	RegisterRoutes(engine)

	return &Server{
		httpServer: &http.Server{
			Addr:              ":" + config.GetConfig(config.Port),
			Handler:           engine,
			ReadHeaderTimeout: 5 * time.Second,
		},
		logger: logger.GetInstance(),
	}
}

// Start runs the server and blocks until SIGINT/SIGTERM, then shuts down gracefully.
func (server *Server) Start() error {
	errChan := make(chan error, 1)
	go func() {
		server.logger.Info(logger.LogMessage{
			Action:  "server.start",
			Message: "Server listening",
			Data:    map[string]any{"addr": server.httpServer.Addr},
		})
		if err := server.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errChan:
		return err
	case <-quit:
	}

	server.logger.Info(logger.LogMessage{Action: "server.shutdown", Message: "Shutting down server"})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return server.httpServer.Shutdown(ctx)
}

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/hasmcp/hasmcp-ce/backend/internal/app"
	zlog "github.com/rs/zerolog/log"
)

const (
	_logPrefix = "[server] "
)

func main() {
	startTime := time.Now().UTC()
	app, err := app.New()

	if err != nil {
		zlog.Fatal().Err(err).Msg(_logPrefix + "failed to init the app")
	}

	var rLimit syscall.Rlimit
	err = syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	ctx := context.Background()

	if err != nil {
		zlog.Warn().Err(err).Msg(_logPrefix + "failed to get rlimit and continuing to start the app")
	}

	zlog.Info().Uint64("current", rLimit.Cur).Uint64("max", rLimit.Max).Msg(_logPrefix + "system ulimits retrieved")

	defer func() {
		if err := recover(); err != nil {
			zlog.Error().Err(fmt.Errorf("%s", err)).Str("stackTrace", string(debug.Stack())).Msg(_logPrefix + "panic recovered")
		}
	}()

	exitCh := make(chan os.Signal, 1)
	signal.Notify(exitCh,
		syscall.SIGTERM, // terminate: stopped by `kill -9 PID`
		syscall.SIGINT,  // interrupt: stopped by Ctrl + C
	)

	go func() {
		defer func() {
			zlog.Info().Msg(_logPrefix + "shutting down the app")
			// send terminate signal when application stop naturally
			exitCh <- syscall.SIGTERM
		}()

		zlog.Info().Str("latency", time.Since(startTime).String()).Msg(_logPrefix + "app start latency")
		err := app.Start(ctx)
		if err != nil {
			zlog.Fatal().Err(err).Msg(_logPrefix + "could not start the app")
			return
		}
	}()

	<-exitCh // blocking until receive exit signal
	stopTime := time.Now().UTC()
	app.Stop(ctx)
	zlog.Info().Str("latency", time.Since(stopTime).String()).Msg(_logPrefix + "app stop latency")
}

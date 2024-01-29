package main

import (
	"errors"
	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"
	"go-apo/anova"
	"log/slog"
	"os"
	"os/signal"
	"reflect"
	"time"
)

func run() error {
	w := os.Stdout
	opts := &tint.Options{
		Level:      slog.LevelDebug,
		TimeFormat: time.DateTime,
		NoColor:    !isatty.IsTerminal(w.Fd()),
	}
	logger := slog.New(tint.NewHandler(w, opts))
	slog.SetDefault(logger)

	refreshToken := os.Getenv("ANOVA_REFRESH_TOKEN")
	if refreshToken == "" {
		return errors.New("missing ANOVA_REFRESH_TOKEN")
	}
	cookerID := anova.CookerID(os.Getenv("ANOVA_COOKER_ID"))
	if cookerID == "" {
		return errors.New("missing ANOVA_COOKER_ID")
	}

	client, err := anova.NewClient(refreshToken, anova.OptionPrintMessageTraces)
	if err != nil {
		return err
	}

	defer client.Close()
	service := anova.NewService(client)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	stop := make(chan bool)
	events := make(chan anova.Event, 1)
	go func() {
		for {
			event := service.ReadEvent()
			switch event.(type) {
			case anova.ServiceStopped:
				stop <- true
				return
			default:
				events <- event
			}
		}
	}()

loop:
	for {
		select {
		case <-interrupt:
			break loop
		case <-stop:
			break loop

		case event := <-events:
			switch event := event.(type) {
			case anova.OvenAdded:
				slog.Info("oven added",
					slog.Any("cookerID", event.Oven.CookerID))

				go func() {
					code, err := event.Oven.GeneratePairingCode()
					if err != nil {
						slog.Error("generate code failed",
							slog.Any("err", err))
					} else {
						slog.Info("generated pairing code",
							slog.String("pairingCode", code))
					}
				}()

			case anova.OvenUpdated:
				slog.Info("oven updated",
					slog.Any("cookerID", event.Oven.CookerID))

			default:
				slog.Debug("ignoring event",
					slog.Any("eventType", reflect.TypeOf(event)))
			}
		}
	}

	slog.Info("shutting down")

	return nil
}

func main() {
	if err := run(); err != nil {
		slog.Error("fatal error",
			slog.Any("err", err))
	}
}

package main

import (
	_ "embed"
	"errors"
	"go-apo/anova"
	"log"
	"os"
	"os/signal"
	"reflect"
)

func run() error {
	refreshToken := os.Getenv("ANOVA_REFRESH_TOKEN")
	if refreshToken == "" {
		return errors.New("missing ANOVA_REFRESH_TOKEN")
	}
	cookerID := anova.CookerID(os.Getenv("ANOVA_COOKER_ID"))
	if cookerID == "" {
		return errors.New("missing ANOVA_COOKER_ID")
	}

	client, err := anova.NewAnovaClient(refreshToken, anova.OptionPrintMessageTraces)
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
				break
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
				log.Printf("oven added: %s\n", event.Oven.CookerID)

			case anova.OvenUpdated:
				log.Printf("oven updated: %s\n", event.Oven.CookerID)

			default:
				log.Printf("ignoring event %s\n", reflect.TypeOf(event))
			}
		}
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatalln("error:", err)
	}
}

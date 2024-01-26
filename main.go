package main

import (
	_ "embed"
	"errors"
	"log"
	"os"
	"os/signal"
)

func run() error {
	refreshToken := os.Getenv("ANOVA_REFRESH_TOKEN")
	if refreshToken == "" {
		return errors.New("missing ANOVA_REFRESH_TOKEN")
	}
	cookerID := CookerID(os.Getenv("ANOVA_COOKER_ID"))
	if cookerID == "" {
		return errors.New("missing ANOVA_COOKER_ID")
	}

	client, err := NewAnovaClient(refreshToken, OptionPrintMessageTraces)
	if err != nil {
		return err
	}

	defer client.Close()

	//command := SetLampCommand{
	//	On: true,
	//}
	//_, err = client.QueueCommand(cookerID, command)
	//if err != nil {
	//	log.Printf("error sending: %s\n", err)
	//}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	select {
	case <-interrupt:
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatalln("error:", err)
	}
}

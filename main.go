package main

import (
	"errors"
	"fmt"
	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"
	"go-apo/anova"
	"go-apo/meatnet"
	"log/slog"
	"os"
	"os/signal"
	"reflect"
	"time"
	"tinygo.org/x/bluetooth"
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

	slog.Info("connecting")

	client, err := anova.NewClient(refreshToken, anova.OptionPrintMessageTraces)
	if err != nil {
		return err
	}

	defer client.Close()
	service := anova.NewService(client)

	slog.Info("service running")

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

				//go func() {
				//	stage1 := anova.NewCookStage(
				//		anova.RackPositionMiddle,
				//		anova.FanSpeedHigh,
				//		anova.NewSousVideSetpointCelsius(70),
				//		anova.RearElementOnly,
				//		anova.NewSteamPercentage(50),
				//		anova.NewTimer(5*time.Minute, anova.TimerTriggerWhenPreheated))
				//	stage2 := anova.NewCookStage(
				//		anova.RackPositionMiddle,
				//		anova.FanSpeedHigh,
				//		anova.NewSousVideSetpointCelsius(100),
				//		anova.RearElementOnly,
				//		anova.NoSteam,
				//		anova.NewTimer(10*time.Minute, anova.TimerTriggerWhenPreheated))
				//	cook, err := anova.NewCook(stage1, stage2)
				//	if err != nil {
				//		slog.Error("create cook failed", slog.Any("err", err))
				//		return
				//	}
				//
				//	slog.Info("starting cook")
				//
				//	err = cook.Start(event.Oven)
				//	if err != nil {
				//		slog.Error("cook start failed", slog.Any("err", err))
				//		return
				//	}
				//
				//	select {
				//	case <-time.After(5 * time.Second):
				//	}
				//
				//	err = cook.Stop()
				//	if err != nil {
				//		slog.Error("stop failed", slog.Any("err", err))
				//		return
				//	}
				//
				//	// require user action -- need to push button, or auto
				//	// rack position
				//	// fan speed
				//	// terminator -- probe, timer, or nothing?
				//	// heating elements -- top and/or bottom and/or rear
				//	// steam -- relative or absolute (or none?)
				//
				//	//oven := event.Oven
				//	//
				//	//command := dto.StartCookCommandV1{
				//	//	CookID: "android-" + uuid.New().String(),
				//	//	Stages: []dto.OvenStage{
				//	//		dto.CookingStage{
				//	//			ID:                 "android-" + uuid.New().String(),
				//	//			Type:               dto.StageTypePreheat,
				//	//			UserActionRequired: false,
				//	//
				//	//			RackPosition: stageRackPositionPtr(dto.StageRackPosition3),
				//	//
				//	//			TemperatureBulbs: &dto.StageTemperatureBulbs{
				//	//				Mode: dto.TemperatureBulbsModeDry,
				//	//				Dry:  dto.TemperatureSetting{Setpoint: dto.NewTemperatureFromCelsius(40)}},
				//	//			HeatingElements: &dto.StageHeatingElements{
				//	//				Top:    dto.HeatingElementSetting{On: false},
				//	//				Bottom: dto.HeatingElementSetting{On: false},
				//	//				Rear:   dto.HeatingElementSetting{On: true},
				//	//			},
				//	//			Fan:  &dto.StageFan{Speed: 100},
				//	//			Vent: &dto.StageVent{Open: false},
				//	//		},
				//	//		dto.CookingStage{
				//	//			ID:                 "android-" + uuid.New().String(),
				//	//			Type:               dto.StageTypeCook,
				//	//			UserActionRequired: false,
				//	//
				//	//			RackPosition: stageRackPositionPtr(dto.StageRackPosition3),
				//	//
				//	//			TemperatureBulbs: &dto.StageTemperatureBulbs{
				//	//				Mode: dto.TemperatureBulbsModeDry,
				//	//				Dry:  dto.TemperatureSetting{Setpoint: dto.NewTemperatureFromCelsius(40)}},
				//	//			HeatingElements: &dto.StageHeatingElements{
				//	//				Top:    dto.HeatingElementSetting{On: false},
				//	//				Bottom: dto.HeatingElementSetting{On: false},
				//	//				Rear:   dto.HeatingElementSetting{On: true},
				//	//			},
				//	//			Fan:  &dto.StageFan{Speed: 100},
				//	//			Vent: &dto.StageVent{Open: false},
				//	//
				//	//			Timer: &dto.StageTimer{Initial: 500},
				//	//		},
				//	//		dto.StopStage{
				//	//			ID:                 "android-" + uuid.New().String(),
				//	//			Type:               dto.StageTypeStop,
				//	//			UserActionRequired: false,
				//	//
				//	//			Timer: &dto.StageTimer{Initial: 600},
				//	//		},
				//	//	},
				//	//}
				//	//_, data, err := client.SendCommand(oven.CookerID, command)
				//	//if err != nil {
				//	//	slog.Error("failed",
				//	//		slog.Any("err", err))
				//	//} else {
				//	//	slog.Info("succeeded",
				//	//		slog.Any("data", data))
				//	//}
				//
				//	//select {
				//	//case <-time.After(5 * time.Second):
				//	//}
				//	//
				//	//err = oven.StopCook()
				//	//if err != nil {
				//	//	slog.Error("stop cook failed",
				//	//		slog.Any("err", err))
				//	//}
				//}()

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

func runMeatNet() error {
	adapter := bluetooth.DefaultAdapter
	err := adapter.Enable()
	if err != nil {
		return err
	}

	service := meatnet.NewService(bluetooth.DefaultAdapter)

	events := make(chan meatnet.Event, 1)
	go func() {
		for {
			event := service.ReadEvent()
			events <- event
		}
	}()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

loop:
	for {
		select {
		case <-interrupt:
			break loop

		case event := <-events:
			switch event := event.(type) {
			case meatnet.ProbeAddedEvent:
				fmt.Printf("probe added %+v\n", *event.Probe)
			case meatnet.ProbeUpdatedEvent:
				fmt.Printf("probe updated %+v\n", *event.Probe)
			case meatnet.DeviceAddedEvent:
				fmt.Printf("device added %+v\n", *event.Device)
			case meatnet.DeviceUpdatedEvent:
				fmt.Printf("device updated %+v\n", *event.Device)

			default:
				fmt.Printf("skipping event %+v %+v\n", reflect.TypeOf(event), event)
			}
		}
	}

	err = service.Stop()
	if err != nil {
		fmt.Printf("service stop failed: %+v\n", err)
	}
	return nil
}

func main() {
	if err := runMeatNet(); err != nil {
		slog.Error("fatal error",
			slog.Any("err", err))
	}
}

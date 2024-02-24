package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Foundation -framework AppKit
#include "ApoCliUrlHandler.h"
*/
import "C"

import (
	"apocli/pkg/anova"
	"apocli/pkg/anova/dto"
	"apocli/pkg/apocli"
	"errors"
	"fmt"
	"github.com/gorilla/schema"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var urlListener = make(chan string, 1)

//export HandleURL
func HandleURL(u *C.char) {
	urlListener <- C.GoString(u)
}

func getOven(cookerID anova.CookerID, stop chan bool, events chan anova.Event) (*anova.Oven, error) {
	stopPolling := make(chan bool, 1)
	oven := make(chan *anova.Oven)
	go func() {
		for {
			select {
			case <-stop:
				return
			case <-stopPolling:
				return
			case event := <-events:
				switch event := event.(type) {
				case anova.OvenAdded:
					// If no cooker ID was specified, grab the first discovered oven
					if cookerID != "" && event.Oven.CookerID != cookerID {
						continue
					}

					oven <- event.Oven
					return
				}
			}
		}
	}()

	select {
	case oven := <-oven:
		return oven, nil
	case <-time.After(5 * time.Second):
		stopPolling <- true
		return nil, fmt.Errorf("timed out waiting for oven")
	}
}

type startCookParams struct {
	temp       anova.TemperatureSetpoint
	elements   anova.HeatingElements
	steam      *anova.SteamPercentage
	fanSpeed   anova.FanSpeed
	terminator anova.StageEndCondition
}

var decoder = schema.NewDecoder()

func parseStartCookParams(paramValues url.Values) (startCookParams, error) {
	params := startCookParams{}

	type rawStartCookParams struct {
		Mode            string  `schema:"mode,required"`
		Temperature     string  `schema:"temp,required"`
		HeatingElements string  `schema:"elements,required"`
		SteamPercentage float64 `schema:"steam"`
		FanSpeed        string  `schema:"speed"`
		Timer           string  `schema:"timer"`
		Trigger         string  `schema:"trigger"`
	}
	rawParams := rawStartCookParams{}

	decoder.IgnoreUnknownKeys(false)
	err := decoder.Decode(&rawParams, paramValues)
	if err != nil {
		return params, errors.New(fmt.Sprintf("decode failed: %+v", err))
	}

	rawParams.Mode = strings.ToLower(rawParams.Mode)
	switch rawParams.Mode {
	case "dry":
		params.temp.Mode = anova.TemperatureModeDry
	case "wet":
		params.temp.Mode = anova.TemperatureModeWet
	default:
		return params, fmt.Errorf("invalid Mode \"%s\": %+v", rawParams.Mode, params)
	}

	tempRegex := regexp.MustCompile(`^([0-9]{1,3}(\.[0-9]*)?)([cfCF])$`)
	tempGroups := tempRegex.FindStringSubmatch(rawParams.Temperature)
	if tempGroups == nil {
		return params, fmt.Errorf("invalid temp \"%s\"", rawParams.Temperature)
	}
	tempValue, err := strconv.ParseFloat(tempGroups[1], 64)
	if err != nil {
		return params, fmt.Errorf("invalid temp value \"%s\"", tempGroups[1])
	}
	switch strings.ToLower(tempGroups[3]) {
	case "c":
		params.temp.TemperatureCelsius = tempValue
	case "f":
		params.temp.TemperatureCelsius = dto.FahrenheitToCelsius(tempValue)
	default:
		return params, fmt.Errorf("invalid temp unit \"%s\"", tempGroups[3])
	}

	for _, element := range strings.Split(rawParams.HeatingElements, ",") {
		switch strings.ToLower(element) {
		case "top":
			params.elements.Top = true
		case "rear":
			params.elements.Rear = true
		case "bottom":
			params.elements.Bottom = true
		default:
			return params, fmt.Errorf("invalid heating element \"%s\"", element)
		}
	}

	if rawParams.SteamPercentage == 0 {
		params.steam = anova.NoSteam
	} else {
		params.steam = anova.NewSteamPercentage(rawParams.SteamPercentage)
	}

	switch strings.ToLower(rawParams.FanSpeed) {
	case "":
		fallthrough
	case "high":
		params.fanSpeed = anova.FanSpeedHigh
	case "medium":
		params.fanSpeed = anova.FanSpeedMedium
	case "low":
		params.fanSpeed = anova.FanSpeedLow
	case "off":
		params.fanSpeed = anova.FanSpeedOff
	default:
		return params, fmt.Errorf("invalid fan speed \"%s\"", rawParams.FanSpeed)
	}

	if rawParams.Timer == "" {
		params.terminator = nil
	} else {
		timerRegex := regexp.MustCompile(`^((\d{1,2})h)?((\d{1,3})m)?((\d{1,2})s)?$`)
		groups := timerRegex.FindStringSubmatch(rawParams.Timer)

		var duration time.Duration
		if rawHours := groups[2]; rawHours != "" {
			hours, err := strconv.ParseInt(rawHours, 10, 32)
			if err != nil {
				return params, fmt.Errorf("invalid hours for Timer \"%s\"", rawParams.Timer)
			}
			duration += time.Hour * time.Duration(hours)
		}
		if rawMinutes := groups[4]; rawMinutes != "" {
			minutes, err := strconv.ParseInt(rawMinutes, 10, 32)
			if err != nil {
				return params, fmt.Errorf("invalid minutes for Timer \"%s\"", rawParams.Timer)
			}
			duration += time.Minute * time.Duration(minutes)
		}
		if rawSeconds := groups[6]; rawSeconds != "" {
			seconds, err := strconv.ParseInt(rawSeconds, 10, 32)
			if err != nil {
				return params, fmt.Errorf("invalid seconds for Timer \"%s\"", rawParams.Timer)
			}
			duration += time.Second * time.Duration(seconds)
		}

		var trigger anova.TimerTrigger
		switch rawParams.Trigger {
		case "":
			fallthrough
		case "preheated":
			trigger = anova.TimerTriggerWhenPreheated
		case "none":
			trigger = anova.TimerTriggerImmediately
		case "button":
			trigger = anova.TimerTriggerManually
		default:
			return params, fmt.Errorf("invalid Trigger \"%s\"", rawParams.Trigger)
		}

		params.terminator = anova.NewTimer(duration, trigger)
	}

	return params, nil
}

func startCook(oven *anova.Oven, rawParams url.Values) error {
	params, err := parseStartCookParams(rawParams)
	if err != nil {
		return fmt.Errorf("failed to parse start-cook params: %+v", err)
	}

	acknowledged := Prompt(fmt.Sprintf("Start cook on oven \"%s\" (%s)?\n\n%+v", oven.Name, oven.CookerID, params))
	if !acknowledged {
		return nil
	}

	stage := anova.NewCookStage(
		anova.RackPositionMiddle,
		params.fanSpeed,
		params.temp,
		params.elements,
		params.steam,
		params.terminator)
	cook, err := anova.NewCook(stage)
	if err != nil {
		return err
	}

	err = cook.Start(oven)
	if err != nil {
		return err
	}

	return nil
}

func stopCook(oven *anova.Oven) error {
	if oven.State.State.Mode != dto.StateModeCook {
		return fmt.Errorf("nothing to stop; oven \"%s\" (%s) is not cooking", oven.Name, oven.CookerID)
	}

	acknowledged := Prompt(fmt.Sprintf("Stop cook on oven \"%s\" (%s)?", oven.Name, oven.CookerID))
	if !acknowledged {
		return nil
	}

	return oven.StopCook()
}

func run() error {
	config, err := apocli.LoadConfig()
	if err != nil {
		return err
	}

	// Wait for the event handler to pass us the URL
	var inputUrl string
	select {
	case inputUrl = <-urlListener:
	case <-time.After(2 * time.Second):
		return fmt.Errorf("timed out waiting for URL")
	}

	parsedUrl, err := url.Parse(inputUrl)
	if err != nil {
		return fmt.Errorf("error parsing url \"%s\": %+v", inputUrl, err)
	}

	// Parse the cooker ID from the hostname, using the default value from the
	// config if required. This leaves the cooker ID set to an empty string if
	// there is no saved default cooker.
	var cookerID anova.CookerID
	rawCookerID := parsedUrl.Hostname()
	if rawCookerID == "default" {
		cookerID = config.DefaultCookerID
	} else {
		found, _ := regexp.MatchString(`^[0-9a-f]{16}$`, rawCookerID)
		if !found {
			return fmt.Errorf("invalid cooker ID \"%s\"", rawCookerID)
		}
		cookerID = anova.CookerID(rawCookerID)
	}

	client, err := anova.NewClient(config.FirebaseRefreshToken, anova.OptionPrintMessageTraces)
	if err != nil {
		return fmt.Errorf("failed to create client: %+v", err)
	}

	defer client.Close()
	service := anova.NewService(client)

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

	oven, err := getOven(cookerID, stop, events)
	if err != nil {
		return err
	}

	params, err := url.ParseQuery(parsedUrl.RawQuery)
	if err != nil {
		return fmt.Errorf("error parsing query string \"%s\": %+v", parsedUrl.RawQuery, err)
	}

	// Route to the command handler
	pathParts := strings.Split(parsedUrl.Path, "/")
	if len(pathParts) != 2 {
		return fmt.Errorf("error parsing path \"%s\"", parsedUrl.Path)
	}
	command := pathParts[1]
	switch command {
	case "start-cook":
		return startCook(oven, params)

	case "stop-cook":
		return stopCook(oven)

	default:
		return fmt.Errorf("unknown command \"%s\"", command)
	}
}

func Prompt(message string) bool {
	return C.ShowAlert(false, C.CString("apocli"), C.CString(message)) == 0
}

func showErrorMessage(message string) {
	C.ShowAlert(true, C.CString("apocli"), C.CString(message))
}

func main() {
	go func() {
		if err := run(); err != nil {
			showErrorMessage(fmt.Sprintf("fatal error: %+v", err))
		}

		os.Exit(1)
	}()

	C.RunApp()
}

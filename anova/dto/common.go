package dto

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
)

type MessageType string

type RequestID string

// A Message is an envelope used for all communications to and from the
// WebSocket server.
type Message struct {
	// Unique ID used for request-response messages. Empty for events.
	RequestID *RequestID  `json:"requestId,omitempty"`
	Command   MessageType `json:"command"`
	Payload   interface{} `json:"payload"`
}

// RawMessage is a copy of Message, but with a string payload to prevent
// the entire payload from being unmarshalled.
type RawMessage struct {
	RequestID *RequestID      `json:"requestId,omitempty"`
	Command   MessageType     `json:"command"`
	Payload   json.RawMessage `json:"payload"`
}

type Temperature struct {
	Celsius float64 `json:"celsius"`
	// Some newer payloads (e.g. SetProbe*) omit the Fahrenheit field entirely.
	// It feels a little unsafe to mark it omitempty for all instances, but
	// here we are.
	Fahrenheit float64 `json:"fahrenheit,omitempty"`
}

func CelsiusToFahrenheit(celsius float64) float64 {
	return math.Round(1.8*celsius + 32)
}

func NewTemperatureFromCelsius(celsius float64) Temperature {
	// The app uses Celsius, then calculates the Fahrenheit value. It's unclear
	// what would happen if these values diverged from each other.
	return Temperature{
		Celsius:    celsius,
		Fahrenheit: CelsiusToFahrenheit(celsius),
	}
}

type StageType string

const (
	StageTypePreheat StageType = "preheat"
	StageTypeCook    StageType = "cook"
	StageTypeStop    StageType = "stop"
)

type StageTimer struct {
	Initial int `json:"initial"`
}

type SteamGeneratorMode string

const (
	SteamGeneratorModeIdle             SteamGeneratorMode = "idle"
	SteamGeneratorModeRelativeHumidity SteamGeneratorMode = "relative-humidity"
	SteamGeneratorModeSteamPercentage  SteamGeneratorMode = "steam-percentage"
)

type SteamSetting struct {
	Setpoint int `json:"setpoint"`
}

type TemperatureUnit string

const (
	TemperatureUnitCelsius    TemperatureUnit = "C"
	TemperatureUnitFahrenheit TemperatureUnit = "F"
)

type Title struct {
	String string
}

func NewTitle(title string) *Title {
	return &Title{String: title}
}

func (title *Title) UnmarshalJSON(data []byte) error {
	// First, try parsing the field as an int, e.g.:
	//   "title": 0,
	var titleInt int
	err := json.Unmarshal(data, &titleInt)
	if err == nil {
		if titleInt == 0 {
			title.String = ""
			return nil
		}
		return fmt.Errorf("unknown title int \"%d\"", titleInt)
	}

	// Otherwise, try parsing the field as a string, e.g.:
	//   "title": "my title",
	var titleStr string
	err = json.Unmarshal(data, &titleStr)
	if err == nil {
		title.String = titleStr
		return nil
	}

	return errors.New("failed to unmarshal title")
}

func (title *Title) MarshalJSON() ([]byte, error) {
	if title.String == "" {
		return json.Marshal(0)
	}
	return json.Marshal(title.String)
}

package dto

import (
	"encoding/json"
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

func TemperatureFromCelsius(celsius float64) Temperature {
	// The app uses Celsius, then calculates the Fahrenheit value. It's unclear
	// what would happen if these values diverged from each other.
	return Temperature{
		Celsius:    celsius,
		Fahrenheit: math.Round(1.8*celsius + 32),
	}
}

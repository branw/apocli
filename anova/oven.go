package anova

import (
	"go-apo/anova/dto"
	"time"
)

type Oven struct {
	// Internally used, immutable ID of the oven
	CookerID CookerID
	Type     dto.CookerType

	// Display name of the oven. Can be changed via the app, or SetName.
	Name string

	State *dto.OvenStateV1

	// Timestamp of the last state update received from this oven. This may be more
	// recent than the last received OvenUpdated event, as the event is not sent for
	// state updates with no changes.
	LastUpdate time.Time

	client *Client
}

func (oven *Oven) SetName(name string) error {
	return oven.client.SetName(oven.CookerID, name)
}

func (oven *Oven) TurnOnLamp(on bool) error {
	return oven.client.SetLamp(oven.CookerID, on)
}

func (oven *Oven) SetLampPreference(on bool) error {
	return oven.client.SetLampPreference(oven.CookerID, on)
}

func (oven *Oven) DisconnectFromAccount() error {
	return oven.client.DisconnectOvenFromAccount(oven.CookerID)
}

func (oven *Oven) GeneratePairingCode() (string, error) {
	return oven.client.GeneratePairingCode(oven.CookerID)
}

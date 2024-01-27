package anova

import (
	"go-apo/anova/dto"
)

type Oven struct {
	client *Client

	CookerID CookerID
	Name     string
	Type     dto.CookerType
	State    *dto.OvenStateV1
}

func (oven *Oven) SetLamp(on bool) error {
	return oven.client.SetLamp(oven.CookerID, on)
}

func (oven *Oven) SetLampPreference(on bool) error {
	return oven.client.SetLampPreference(oven.CookerID, on)
}

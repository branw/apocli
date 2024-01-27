package anova

import "go-apo/anova/dto"

type Event interface {
	isEvent()
}

type OvenAdded struct {
	Oven *Oven
}

func (OvenAdded) isEvent() {}

type OvenUpdated struct {
	Oven          *Oven
	PreviousState *dto.OvenStateV1
}

func (OvenUpdated) isEvent() {}

type ServiceStopped struct{}

func (ServiceStopped) isEvent() {}

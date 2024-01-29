package anova

import "go-apo/anova/dto"

type Event interface {
	isEvent()
}

// OvenAdded indicates that a new oven has been loaded by the service. This event
// is sent the first time an oven is found by the service, which can happen
// initially at startup, or when an oven is newly paired to the account.
type OvenAdded struct {
	Oven *Oven
}

func (OvenAdded) isEvent() {}

// OvenRenamed indicates that a known oven has a new name.
type OvenRenamed struct {
	Oven    *Oven
	OldName string
}

func (OvenRenamed) isEvent() {}

// OvenUpdated indicates that the oven state has materially changed. This event is
// only sent if the oven's state changes after the initial OvenAdded event.
type OvenUpdated struct {
	Oven          *Oven
	PreviousState *dto.OvenStateV1
}

func (OvenUpdated) isEvent() {}

// ServiceStopped indicates that the client connection has been closed and no
// more events can be processed.
type ServiceStopped struct{}

func (ServiceStopped) isEvent() {}

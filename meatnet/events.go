package meatnet

type Event interface {
	isEvent()
}

type ProbeAddedEvent struct {
	Probe *Probe
}

func (event ProbeAddedEvent) isEvent() {}

type ProbeUpdatedEvent struct {
	Probe *Probe
}

func (event ProbeUpdatedEvent) isEvent() {}

type DeviceAddedEvent struct {
	Device *Device
}

func (event DeviceAddedEvent) isEvent() {}

type DeviceUpdatedEvent struct {
	Device *Device
}

func (event DeviceUpdatedEvent) isEvent() {}

package meatnet

import (
	"fmt"
	"github.com/r3labs/diff/v3"
	"go-apo/bitbuffer"
	"reflect"
	"time"
	"tinygo.org/x/bluetooth"
)

type Probe struct {
	SerialNumber uint32
	LastUpdated  time.Time

	TemperatureReadings [8]float64

	LastInstantReading time.Time
	InstantReading     float64

	VirtualCoreSensor    Sensor
	VirtualSurfaceSensor Sensor
	VirtualAmbientSensor Sensor

	ColorID ColorID
	ProbeID ProbeID

	BatteryStatus BatteryStatus
}

func (probe Probe) String() string {
	return fmt.Sprintf("Probe %08X {Temperatures:[%.02f %.02f %.02f %.02f %.02f %.02f %.02f %.02f] Instant:%.02f Core:%s Surface:%s Ambient:%s ColorID:%s ProbeID:%d}",
		probe.SerialNumber,
		probe.TemperatureReadings[0], probe.TemperatureReadings[1], probe.TemperatureReadings[2], probe.TemperatureReadings[3], probe.TemperatureReadings[4], probe.TemperatureReadings[5], probe.TemperatureReadings[6], probe.TemperatureReadings[7],
		probe.InstantReading,
		probe.VirtualCoreSensor, probe.VirtualSurfaceSensor, probe.VirtualAmbientSensor,
		probe.ColorID, probe.ProbeID)
}

type Device struct {
	Address     bluetooth.Address
	LastUpdated time.Time

	Type ProductType

	ReportedProbes map[uint32]*Probe
}

type Service struct {
	adapter *bluetooth.Adapter

	events chan Event

	Devices map[bluetooth.Address]*Device
	Probes  map[uint32]*Probe
}

func NewService(adapter *bluetooth.Adapter) *Service {
	service := &Service{
		adapter: adapter,

		events: make(chan Event),

		Devices: make(map[bluetooth.Address]*Device),
		Probes:  make(map[uint32]*Probe),
	}

	go service.receiveAdvertisements()

	return service
}

func (service *Service) Stop() error {
	return service.adapter.StopScan()
}

func (service *Service) ReadEvent() Event {
	select {
	case event := <-service.events:
		return event
	}
}

func (service *Service) writeEvent(event Event) {
	select {
	case service.events <- event:
	}
}

func (service *Service) receiveAdvertisements() {
	err := service.adapter.Scan(func(adapter *bluetooth.Adapter, device bluetooth.ScanResult) {
		data, exists := device.ManufacturerData()[0x09C7]
		if !exists {
			return
		}

		var manufacturerData ManufacturerData
		err := bitbuffer.UnmarshalFully(data, &manufacturerData)
		if err != nil {
			fmt.Printf("failed to parse manufacturer data: %+v\n", err)
			return
		}

		if manufacturerData.ProductType != ProductTypePredictiveProbe &&
			manufacturerData.ProductType != ProductTypeRepeaterNode {
			fmt.Printf("unknown product type: %+v\n", manufacturerData.ProductType)
			return
		}

		timestamp := time.Now()
		probe := service.updateProbe(manufacturerData, timestamp)
		service.updateDevice(device, manufacturerData, probe, timestamp)
	})
	if err != nil {
		fmt.Printf("adapter scan failed: %+v\n", err)
	}
}

func (service *Service) updateProbe(data ManufacturerData, timestamp time.Time) *Probe {
	probe, probePreviouslyAdded := service.Probes[data.SerialNumber]
	if !probePreviouslyAdded {
		probe = &Probe{
			SerialNumber: data.SerialNumber,
		}
		service.Probes[probe.SerialNumber] = probe
	}

	previousProbe := *probe

	if data.Mode == ModeNormal {
		for i := 0; i < 8; i++ {
			probe.TemperatureReadings[i] = data.RawThermistorData[i].Celsius()
		}
	} else if data.Mode == ModeInstantRead {
		probe.InstantReading = data.RawThermistorData[0].Celsius()
		probe.LastInstantReading = timestamp
	}

	probe.VirtualCoreSensor = data.VirtualCoreSensor.Sensor()
	probe.VirtualSurfaceSensor = data.VirtualSurfaceSensor.Sensor()
	probe.VirtualAmbientSensor = data.VirtualAmbientSensor.Sensor()

	probe.ColorID = data.ColorID
	probe.ProbeID = data.ProbeID

	probe.BatteryStatus = data.BatteryStatus

	probe.LastUpdated = timestamp

	if !probePreviouslyAdded {
		service.writeEvent(ProbeAddedEvent{Probe: probe})
		return probe
	}

	diffFilter := diff.Filter(func(path []string, parent reflect.Type, field reflect.StructField) bool {
		return !(len(path) >= 1 && path[0] == "LastUpdated")
	})
	changelog, err := diff.Diff(previousProbe, *probe, diffFilter)
	if err != nil {
		fmt.Printf("diff failed: %+v\n", err)
	}
	if len(changelog) > 0 {
		service.writeEvent(ProbeUpdatedEvent{Probe: probe})
	}
	return probe
}

func (service *Service) updateDevice(device bluetooth.ScanResult, data ManufacturerData, probe *Probe, timestamp time.Time) *Device {
	meatNetDevice, devicePreviouslyAdded := service.Devices[device.Address]
	if !devicePreviouslyAdded {
		meatNetDevice = &Device{
			Address: device.Address,

			Type: data.ProductType,

			ReportedProbes: make(map[uint32]*Probe),
		}
		service.Devices[device.Address] = meatNetDevice
	}

	previousDevice := *meatNetDevice

	if _, exists := meatNetDevice.ReportedProbes[data.SerialNumber]; !exists {
		meatNetDevice.ReportedProbes[data.SerialNumber] = probe
	}

	meatNetDevice.LastUpdated = timestamp

	if !devicePreviouslyAdded {
		service.writeEvent(DeviceAddedEvent{Device: meatNetDevice})
		return meatNetDevice
	}

	diffFilter := diff.Filter(func(path []string, parent reflect.Type, field reflect.StructField) bool {
		return !(len(path) >= 1 && path[0] == "LastUpdated")
	})
	changelog, err := diff.Diff(previousDevice, *meatNetDevice, diffFilter)
	if err != nil {
		fmt.Printf("diff failed: %+v\n", err)
	}
	if len(changelog) > 0 {
		service.writeEvent(DeviceUpdatedEvent{Device: meatNetDevice})
	}
	return meatNetDevice
}

package meatnet

import (
	"errors"
	"fmt"
	"go-apo/bitbuffer"
	"reflect"
)

type ProductType uint8

const (
	ProductTypeUnknown         ProductType = 0
	ProductTypePredictiveProbe ProductType = 1
	ProductTypeRepeaterNode    ProductType = 2
)

func (pt ProductType) Validate() error {
	if pt != ProductTypeUnknown && pt != ProductTypePredictiveProbe && pt != ProductTypeRepeaterNode {
		return errors.New("invalid product type")
	}
	return nil
}

type Mode uint8

const (
	ModeNormal      Mode = 0
	ModeInstantRead Mode = 1
	ModeReserved    Mode = 2
	ModeError       Mode = 3
)

func (m Mode) String() string {
	switch m {
	case ModeNormal:
		return "normal"
	case ModeInstantRead:
		return "instant-read"
	case ModeReserved:
		return "reserved"
	case ModeError:
		return "error"
	default:
		return "<unknown mode>"
	}
}

func (m Mode) Validate() error {
	if m != ModeNormal && m != ModeInstantRead && m != ModeReserved && m != ModeError {
		return errors.New("invalid mode")
	}
	return nil
}

type ColorID uint8

const (
	ColorYellow ColorID = 0
	ColorGrey   ColorID = 1
)

func (c ColorID) String() string {
	switch c {
	case ColorYellow:
		return "yellow"
	case ColorGrey:
		return "grey"
	default:
		return "<unknown color>"
	}
}

type ProbeID uint8

type BatteryStatus uint8

const (
	BatteryStatusOK  BatteryStatus = 0
	BatteryStatusLow BatteryStatus = 1
)

func (bs BatteryStatus) String() string {
	switch bs {
	case BatteryStatusOK:
		return "ok"
	case BatteryStatusLow:
		return "low-battery"
	default:
		return "<unknown battery status>"
	}
}

func (bs BatteryStatus) Validate() error {
	if bs != BatteryStatusOK && bs != BatteryStatusLow {
		return errors.New("invalid battery status")
	}
	return nil
}

type RawThermistorValue uint16

func (rawValue RawThermistorValue) Celsius() float64 {
	return (float64(rawValue) * 0.05) - 20
}

type RawThermistorValues [8]RawThermistorValue

func (data *RawThermistorValues) Unmarshal(bb *bitbuffer.BitBuffer, _ reflect.Value, _ reflect.StructTag) error {
	for i := 0; i < 8; i++ {
		thermistorValue, err := bb.ReadBits(13)
		if err != nil {
			return err
		}
		data[i] = RawThermistorValue(thermistorValue)
	}
	return nil
}

type Sensor uint8

const (
	SensorUnknown Sensor = 0
	SensorT1      Sensor = 1
	SensorT2      Sensor = 2
	SensorT3      Sensor = 3
	SensorT4      Sensor = 4
	SensorT5      Sensor = 5
	SensorT6      Sensor = 6
	SensorT7      Sensor = 7
	SensorT8      Sensor = 8
)

func (s Sensor) String() string {
	return fmt.Sprintf("T%d", s)
}

type VirtualCoreSensor uint8

const (
	VirtualCoreSensorT1 VirtualCoreSensor = 0
	VirtualCoreSensorT2 VirtualCoreSensor = 1
	VirtualCoreSensorT3 VirtualCoreSensor = 2
	VirtualCoreSensorT4 VirtualCoreSensor = 3
	VirtualCoreSensorT5 VirtualCoreSensor = 4
	VirtualCoreSensorT6 VirtualCoreSensor = 5
)

func (sensor VirtualCoreSensor) Sensor() Sensor {
	switch sensor {
	case VirtualCoreSensorT1:
		return SensorT1
	case VirtualCoreSensorT2:
		return SensorT2
	case VirtualCoreSensorT3:
		return SensorT3
	case VirtualCoreSensorT4:
		return SensorT4
	case VirtualCoreSensorT5:
		return SensorT5
	case VirtualCoreSensorT6:
		return SensorT6
	default:
		return SensorUnknown
	}
}

type VirtualSurfaceSensor uint8

const (
	VirtualSurfaceSensorT4 VirtualSurfaceSensor = 0
	VirtualSurfaceSensorT5 VirtualSurfaceSensor = 1
	VirtualSurfaceSensorT6 VirtualSurfaceSensor = 2
	VirtualSurfaceSensorT7 VirtualSurfaceSensor = 3
)

func (sensor VirtualSurfaceSensor) Sensor() Sensor {
	switch sensor {
	case VirtualSurfaceSensorT4:
		return SensorT4
	case VirtualSurfaceSensorT5:
		return SensorT5
	case VirtualSurfaceSensorT6:
		return SensorT6
	case VirtualSurfaceSensorT7:
		return SensorT7
	default:
		return SensorUnknown
	}
}

type VirtualAmbientSensor uint8

const (
	VirtualAmbientSensorT5 VirtualAmbientSensor = 0
	VirtualAmbientSensorT6 VirtualAmbientSensor = 1
	VirtualAmbientSensorT7 VirtualAmbientSensor = 2
	VirtualAmbientSensorT8 VirtualAmbientSensor = 3
)

func (sensor VirtualAmbientSensor) Sensor() Sensor {
	switch sensor {
	case VirtualAmbientSensorT5:
		return SensorT5
	case VirtualAmbientSensorT6:
		return SensorT6
	case VirtualAmbientSensorT7:
		return SensorT7
	case VirtualAmbientSensorT8:
		return SensorT8
	default:
		return SensorUnknown
	}
}

type ManufacturerData struct {
	ProductType  ProductType `bbwidth:"8"`
	SerialNumber uint32

	RawThermistorData RawThermistorValues

	Mode    Mode    `bbwidth:"2"`
	ColorID ColorID `bbwidth:"3"`
	ProbeID ProbeID `bbwidth:"3"`

	BatteryStatus        BatteryStatus        `bbwidth:"1"`
	VirtualCoreSensor    VirtualCoreSensor    `bbwidth:"3"`
	VirtualSurfaceSensor VirtualSurfaceSensor `bbwidth:"2"`
	VirtualAmbientSensor VirtualAmbientSensor `bbwidth:"2"`

	HopCount                   uint8 `bbwidth:"2"`
	ReservedNetworkInformation uint8 `bbwidth:"6"`

	Reserved uint8 `bbwidth:"8"`
}

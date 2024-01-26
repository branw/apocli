package main

import (
	"encoding/json"
	"math"
	"reflect"
	_ "reflect"
	"time"
)

type MessageType string

type RequestID string

var (
	messageTypeToResponseType = map[MessageType]reflect.Type{
		"EVENT_APO_WIFI_LIST": reflect.TypeOf(&EventApoWifiList{}).Elem(),
		"EVENT_USER_STATE":    reflect.TypeOf(&EventUserState{}).Elem(),
		"EVENT_APO_STATE":     reflect.TypeOf(&EventApoState{}).Elem(),
		"RESPONSE":            reflect.TypeOf(&Response{}).Elem(),
	}
)

var (
	requestTypeToMessageType = map[reflect.Type]MessageType{
		reflect.TypeOf(&SetFanCommand{}).Elem():              "",
		reflect.TypeOf(&SetHeatingElementsCommand{}).Elem():  "",
		reflect.TypeOf(&SetLampCommand{}).Elem():             "CMD_APO_SET_LAMP",
		reflect.TypeOf(&SetLampPreferenceCommand{}).Elem():   "",
		reflect.TypeOf(&SetProbeCommand{}).Elem():            "",
		reflect.TypeOf(&SetSteamGeneratorsCommand{}).Elem():  "",
		reflect.TypeOf(&SetTemperatureBulbsCommand{}).Elem(): "",
		reflect.TypeOf(&SetTemperatureUnitCommand{}).Elem():  "",
		reflect.TypeOf(&SetTimerCommand{}).Elem():            "",
		reflect.TypeOf(&SetVentCommand{}).Elem():             "",
		reflect.TypeOf(&StartCookCommand{}).Elem():           "",
		reflect.TypeOf(&StartDescaleCommand{}).Elem():        "",
		reflect.TypeOf(&AbortDescaleCommand{}).Elem():        "",
		reflect.TypeOf(&StartFirmwareUpdateCommand{}).Elem(): "",
		reflect.TypeOf(&StartStageCommand{}).Elem():          "",
		reflect.TypeOf(&StopCookCommand{}).Elem():            "",
		reflect.TypeOf(&UpdateCookStageCommand{}).Elem():     "",
		reflect.TypeOf(&UpdateCookStagesCommand{}).Elem():    "",
		reflect.TypeOf(&UpdateCookStagesV1Command{}).Elem():  "",
		reflect.TypeOf(&SetReportStateRate{}).Elem():         "",
		reflect.TypeOf(&SetReportStateRateDefault{}).Elem():  "",
		reflect.TypeOf(&SetBoilerTime{}).Elem():              "CMD_APO_SET_BOILER_TIME",
		reflect.TypeOf(&AuthTokenV2{}).Elem():                "",
		reflect.TypeOf(&HealthCheck{}).Elem():                "",
		reflect.TypeOf(&SetMetadataCommand{}).Elem():         "CMD_APO_SET_METADATA",
		reflect.TypeOf(&GetConfiguration{}).Elem():           "",
		reflect.TypeOf(&SetConfiguration{}).Elem():           "",
		reflect.TypeOf(&NameWifiDeviceCommand{}).Elem():      "CMD_APO_NAME_WIFI_DEVICE",
		reflect.TypeOf(&DisconnectCommand{}).Elem():          "",
		reflect.TypeOf(&RequestDiagnosticsCommand{}).Elem():  "",
		reflect.TypeOf(&RegisterPushTokenCommand{}).Elem():   "",
		reflect.TypeOf(&SetTimeZoneCommand{}).Elem():         "",
		reflect.TypeOf(&StartLiveStreamCommand{}).Elem():     "",
		reflect.TypeOf(&StopLiveStreamCommand{}).Elem():      "",
	}
)

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

/*
 * Request objects (sent to server)
 */

type CookerID string

type Command struct {
	ID      CookerID    `json:"id"`
	Type    MessageType `json:"type"`
	Payload interface{} `json:"payload"`
}

type SetFanCommand struct {
}

type SetHeatingElementsCommand struct {
}

type SetLampCommand struct {
	On bool `json:"on"`
}

type SetLampPreferenceCommand struct {
}

type SetProbeCommand struct {
}

type SetSteamGeneratorsCommand struct {
}

type SetTemperatureBulbsCommand struct {
}

type SetTemperatureUnitCommand struct {
}

type SetTimerCommand struct {
}

type SetVentCommand struct {
}

type StartCookCommand struct {
}

type StartDescaleCommand struct {
}

type AbortDescaleCommand struct {
}

type StartFirmwareUpdateCommand struct {
}

type StartStageCommand struct {
}

type StopCookCommand struct {
}

type UpdateCookStageCommand struct {
}

type UpdateCookStagesCommand struct {
}

type UpdateCookStagesV1Command struct {
}

type SetReportStateRate struct {
}

type SetReportStateRateDefault struct {
}

type SetBoilerTime struct {
	Time int `json:"time"`
}

type AuthTokenV2 struct {
}

type HealthCheck struct {
}

type SetMetadataCommand struct {
	Metadata map[string]interface{} `json:"metadata"`
}

type GetConfiguration struct {
}

type SetConfiguration struct {
}

type NameWifiDeviceCommand struct {
	Name string `json:"name"`
}

type DisconnectCommand struct {
}

type RequestDiagnosticsCommand struct {
}

type RegisterPushTokenCommand struct {
}

type SetTimeZoneCommand struct {
}

type StartLiveStreamCommand struct {
}

type StopLiveStreamCommand struct {
}

/*
 * Response objects (received from server)
 */

// Only implemented for "oven_v1"
// TODO Implement "oven_v2" structures
type EventApoState struct {
	CookerID CookerID    `json:"cookerId"`
	Type     string      `json:"type"`
	State    OvenStateV1 `json:"state"`
}

type ResponseStatus string

const (
	ResponseStatusOk    ResponseStatus = "ok"
	ResponseStatusError ResponseStatus = "error"
)

type Response struct {
	Status ResponseStatus `json:"status"`
	Error  *string        `json:"error,omitempty"`
}

type EventApoWifiList []struct {
	CookerID CookerID `json:"cookerId"`
	Name     string   `json:"name"`
	PairedAt string   `json:"pairedAt"`
	Type     string   `json:"type"`
}

type EventUserState struct {
	IsConnectedToAlexa      bool `json:"isConnectedToAlexa"`
	IsConnectedToGoogleHome bool `json:"isConnectedToGoogleHome"`
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

type StageFan struct {
	Speed int `json:"speed"`
}

type StageHeatingElements struct {
	Bottom struct {
		On bool `json:"on"`
	} `json:"bottom"`
	Rear struct {
		On bool `json:"on"`
	} `json:"rear"`
	Top struct {
		On bool `json:"on"`
	} `json:"top"`
}

type StageRackPosition int

const (
	StageRackPosition1 StageRackPosition = 1
	StageRackPosition2 StageRackPosition = 2
	StageRackPosition3 StageRackPosition = 3
	StageRackPosition4 StageRackPosition = 4
	StageRackPosition5 StageRackPosition = 5
)

// TODO Maybe merge with SteamGeneratorsNode?
type StageSteamGenerators struct {
	Mode SteamGeneratorMode `json:"mode"`

	RelativeHumidity struct {
		Setpoint int `json:"setpoint"`
	} `json:"relativeHumidity,omitempty"`
}

type StageTemperatureBulbs struct {
	Mode TemperatureBulbsMode `json:"mode"`
	Dry  struct {
		Setpoint Temperature `json:"setpoint"`
	} `json:"dry,omitempty"`
	Wet struct {
		Setpoint Temperature `json:"setpoint"`
	} `json:"wet,omitempty"`
}

type StageTemperatureProbe struct {
	Setpoint Temperature `json:"setpoint"`
}

type StageVent struct {
	Open bool `json:"open"`
}

// "CookingStage" and "StopStage" combined
type Stage struct {
	ID                 string    `json:"id"`
	Type               StageType `json:"type"`
	UserActionRequired bool      `json:"userActionRequired"`

	Description *string     `json:"description,omitempty"`
	Timer       *StageTimer `json:"timer,omitempty"`
	// integer 0 for no title, populated string for title
	Title *interface{} `json:"title,omitempty"`

	Fan              *StageFan              `json:"fan"`
	HeatingElements  *StageHeatingElements  `json:"heatingElements,omitempty"`
	RackPosition     *StageRackPosition     `json:"rackPosition,omitempty"`
	SteamGenerators  *StageSteamGenerators  `json:"steamGenerators,omitempty"`
	TemperatureBulbs *StageTemperatureBulbs `json:"temperatureBulbs,omitempty"`
	ProbeAdded       *bool                  `json:"probeAdded,omitempty"`
	TemperatureProbe *StageTemperatureProbe `json:"temperatureProbe,omitempty"`
	Vent             *StageVent             `json:"vent,omitempty"`

	PhotoUrl          *string `json:"photoUrl,omitempty"`
	VideoUrl          *string `json:"videoUrl,omitempty"`
	VideoThumbnailUrl *string `json:"videoThumbnailUrl,omitempty"`
}

type CookV1 struct {
	ActiveStageID                    string  `json:"activeStageId"`
	ActiveStageIndex                 int     `json:"activeStageIndex"`
	ActiveStageSecondsElapsed        int     `json:"activeStageSecondsElapsed"`
	CookID                           string  `json:"cookId"`
	SecondsElapsed                   int     `json:"secondsElapsed"`
	StageTransitionPendingUserAction bool    `json:"stageTransitionPendingUserAction"`
	Stages                           []Stage `json:"stages"`
}

type DoorNode struct {
	Closed bool `json:"closed"`
}

type FanNode struct {
	Speed  int  `json:"speed"`
	Failed bool `json:"failed"`
}

type HeatingElement struct {
	On     bool `json:"on"`
	Failed bool `json:"failed"`
	Watts  int  `json:"watts"`
}

type HeatingElementsNode struct {
	Top    HeatingElement `json:"top"`
	Bottom HeatingElement `json:"bottom"`
	Rear   HeatingElement `json:"rear"`
}

type LampNode struct {
	On         bool   `json:"on"`
	Failed     bool   `json:"failed"`
	Preference string `json:"preference"`
}

type EvaporatorSteamGeneratorNode struct {
	Failed     bool    `json:"failed"`
	Overheated bool    `json:"overheated"`
	Celsius    float64 `json:"celsius"`
	Watts      int     `json:"watts"`
}

type BoilerSteamGeneratorNode struct {
	DescaleRequired bool    `json:"descaleRequired"`
	Failed          bool    `json:"failed"`
	Overheated      bool    `json:"overheated"`
	Celsius         float64 `json:"celsius"`
	Watts           int     `json:"watts"`
	Dosed           bool    `json:"dosed"`
}

type RelativeHumiditySteamGeneratorNode struct {
	Current  int `json:"current"`
	Setpoint int `json:"setpoint"`
}

type SteamPercentageSteamGeneratorNode struct {
	Setpoint int `json:"setpoint"`
}

type SteamGeneratorMode string

const (
	SteamGeneratorModeIdle             SteamGeneratorMode = "idle"
	SteamGeneratorModeRelativeHumidity SteamGeneratorMode = "relative-humidity"
	SteamGeneratorModeSteamPercentage  SteamGeneratorMode = "steam-percentage"
)

type SteamGeneratorsNode struct {
	Mode SteamGeneratorMode `json:"mode"`

	RelativeHumidity *RelativeHumiditySteamGeneratorNode `json:"relativeHumidity,omitempty"`
	SteamPercentage  *SteamPercentageSteamGeneratorNode  `json:"steamPercentage,omitempty"`

	Evaporator EvaporatorSteamGeneratorNode `json:"evaporator"`
	Boiler     BoilerSteamGeneratorNode     `json:"boiler"`
}

type Temperature struct {
	Celsius    float64 `json:"celsius"`
	Fahrenheit float64 `json:"fahrenheit"`
}

func TemperatureFromCelsius(celsius float64) Temperature {
	// The app uses Celsius, then calculates the Fahrenheit value. It's unclear
	// what would happen if these values diverged from each other.
	return Temperature{
		Celsius:    celsius,
		Fahrenheit: math.Round(1.8*celsius + 32),
	}
}

type TemperatureBulbsMode string

const (
	TemperatureBulbsModeWet TemperatureBulbsMode = "wet"
	TemperatureBulbsModeDry TemperatureBulbsMode = "dry"
)

type TemperatureBulbsNodeV1 struct {
	Mode TemperatureBulbsMode `json:"mode"`
	Wet  struct {
		Current Temperature `json:"current"`
		// Only present in "wet" mode
		Setpoint   *Temperature `json:"setpoint,omitempty"`
		Dosed      bool         `json:"dosed"`
		DoseFailed bool         `json:"doseFailed"`
	} `json:"wet"`
	Dry struct {
		Current Temperature `json:"current"`
		// Only present in "dry" mode
		Setpoint *Temperature `json:"setpoint,omitempty"`
	} `json:"dry"`
	DryTop struct {
		Current    Temperature `json:"current"`
		Overheated bool        `json:"overheated"`
	} `json:"dryTop"`
	DryBottom struct {
		Current    Temperature `json:"current"`
		Overheated bool        `json:"overheated"`
	} `json:"dryBottom"`
}

type TemperatureProbeNodeV1 struct {
	Connected bool `json:"connected"`
}

type TimerMode string

const (
	TimerModeIdle    TimerMode = "idle"
	TimerModePaused  TimerMode = "paused"
	TimerModeRunning TimerMode = "running"
)

type TimerNodeV1 struct {
	Mode    string `json:"mode"`
	Initial int    `json:"initial"`
	Current int    `json:"current"`
}

type UserInterfaceCircuitNode struct {
	CommunicationFailed bool `json:"communicationFailed"`
}

type VentNode struct {
	Open bool `json:"open"`
}

type WaterTankNodeV1 struct {
	Empty bool `json:"empty"`
}

type NodesV1 struct {
	Door                 DoorNode                 `json:"door"`
	Fan                  FanNode                  `json:"fan"`
	HeatingElements      HeatingElementsNode      `json:"heatingElements"`
	Lamp                 LampNode                 `json:"lamp"`
	SteamGenerators      SteamGeneratorsNode      `json:"steamGenerators"`
	TemperatureBulbs     TemperatureBulbsNodeV1   `json:"temperatureBulbs"`
	Timer                TimerNodeV1              `json:"timer"`
	TemperatureProbe     TemperatureProbeNodeV1   `json:"temperatureProbe"`
	UserInterfaceCircuit UserInterfaceCircuitNode `json:"userInterfaceCircuit"`
	Vent                 VentNode                 `json:"vent"`
	WaterTank            WaterTankNodeV1          `json:"waterTank"`
}

type StateMode string

const (
	StateModeIdle    StateMode = "idle"
	StateModeCook    StateMode = "cook"
	StateModeDescale StateMode = "descale"
)

type TemperatureUnit string

const (
	TemperatureUnitCelsius    TemperatureUnit = "C"
	TemperatureUnitFahrenheit TemperatureUnit = "F"
)

type StateV1 struct {
	Mode            StateMode       `json:"mode"`
	TemperatureUnit TemperatureUnit `json:"temperatureUnit"`
	// List of request IDs for the 10 most recently processed commands
	ProcessedCommandIDs []string `json:"processedCommandIds"`
}

type OtaUpdateMode string

const (
	OtaUpdateModeDefault            OtaUpdateMode = "default"
	OtaUpdateModeDownloading        OtaUpdateMode = "downloading"
	OtaUpdateModeDownloadingPartial OtaUpdateMode = "downloading_partial"
	OtaUpdateModeError              OtaUpdateMode = "error"
	OtaUpdateModePowerBoardUpdating OtaUpdateMode = "power_board_updating"
	OtaUpdateModeReboot             OtaUpdateMode = "reboot"
	OtaUpdateModeRollback           OtaUpdateMode = "rollback"
	OtaUpdateModeUpdate             OtaUpdateMode = "update"
	OtaUpdateModeUpdatePartial      OtaUpdateMode = "update_partial"
)

type OtaUpdateV1 struct {
	Mode OtaUpdateMode
}

type SystemInfoV1 struct {
	BetaFeature               *string      `json:"betaFeature,omitempty"`
	Online                    bool         `json:"online"`
	HardwareVersion           string       `json:"hardwareVersion"`
	PowerMains                int          `json:"powerMains"`
	PowerHertz                int          `json:"powerHertz"`
	FirmwareVersion           string       `json:"firmwareVersion"`
	OtaUpdate                 *OtaUpdateV1 `json:"otaUpdate,omitempty"`
	UIHardwareVersion         *string      `json:"uiHardwareVersion,omitempty"`
	UIFirmwareVersion         *string      `json:"uiFirmwareVersion,omitempty"`
	FirmwareUpdatedTimestamp  time.Time    `json:"firmwareUpdatedTimestamp"`
	LastConnectedTimestamp    time.Time    `json:"lastConnectedTimestamp"`
	LastDisconnectedTimestamp time.Time    `json:"lastDisconnectedTimestamp"`
	TriacsFailed              bool         `json:"triacsFailed"`
}

type OvenStateV1 struct {
	Cook             *CookV1      `json:"cook,omitempty"`
	Nodes            NodesV1      `json:"nodes"`
	State            StateV1      `json:"state"`
	SystemInfo       SystemInfoV1 `json:"systemInfo"`
	Version          int          `json:"version"`
	UpdatedTimestamp time.Time    `json:"updatedTimestamp"`
}

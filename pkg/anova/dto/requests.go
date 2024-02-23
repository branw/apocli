package dto

import (
	"reflect"
)

var (
	RequestTypeToMessageType = map[reflect.Type]MessageType{
		reflect.TypeOf(&AuthTokenV2{}).Elem():                "AUTH_TOKEN_V2",
		reflect.TypeOf(&HealthCheck{}).Elem():                "CMD_APO_HEALTHCHECK",
		reflect.TypeOf(&DisconnectCommand{}).Elem():          "CMD_APO_DISCONNECT",
		reflect.TypeOf(&RegisterPushTokenCommand{}).Elem():   "CMD_APO_REGISTER_PUSH_TOKEN",
		reflect.TypeOf(&NameWifiDeviceCommand{}).Elem():      "CMD_APO_NAME_WIFI_DEVICE",
		reflect.TypeOf(&StartFirmwareUpdateCommand{}).Elem(): "CMD_APO_OTA",
		reflect.TypeOf(&GetConfiguration{}).Elem():           "CMD_APO_GET_CONFIGURATION",
		reflect.TypeOf(&SetConfiguration{}).Elem():           "CMD_APO_SET_CONFIGURATION",
		reflect.TypeOf(&SetReportStateRate{}).Elem():         "CMD_APO_SET_REPORT_STATE_RATE",
		reflect.TypeOf(&SetReportStateRateDefault{}).Elem():  "CMD_APO_SET_REPORT_STATE_RATE_DEFAULT",
		reflect.TypeOf(&SetMetadataCommand{}).Elem():         "CMD_APO_SET_METADATA",
		reflect.TypeOf(&SetTimeZoneCommand{}).Elem():         "CMD_APO_SET_TIME_ZONE",
		reflect.TypeOf(&RequestDiagnosticsCommand{}).Elem():  "CMD_APO_REQUEST_DIAGNOSTIC",

		reflect.TypeOf(&SetBoilerTime{}).Elem():              "CMD_APO_SET_BOILER_TIME",
		reflect.TypeOf(&SetFanCommand{}).Elem():              "CMD_APO_SET_FAN",
		reflect.TypeOf(&SetHeatingElementsCommand{}).Elem():  "CMD_APO_SET_HEATING_ELEMENTS",
		reflect.TypeOf(&SetLampCommand{}).Elem():             "CMD_APO_SET_LAMP",
		reflect.TypeOf(&SetLampPreferenceCommand{}).Elem():   "CMD_APO_SET_LAMP_PREFERENCE",
		reflect.TypeOf(&SetProbeCommand{}).Elem():            "CMD_APO_SET_PROBE",
		reflect.TypeOf(&SetSteamGeneratorsCommand{}).Elem():  "CMD_APO_SET_STEAM_GENERATORS",
		reflect.TypeOf(&SetTemperatureBulbsCommand{}).Elem(): "CMD_APO_SET_TEMPERATURE_BULBS",
		reflect.TypeOf(&SetTemperatureUnitCommand{}).Elem():  "CMD_APO_SET_TEMPERATURE_UNIT",
		reflect.TypeOf(&SetTimerCommand{}).Elem():            "CMD_APO_SET_TIMER",
		reflect.TypeOf(&SetVentCommand{}).Elem():             "CMD_APO_SET_VENT",

		reflect.TypeOf(&StartCookCommandV1{}).Elem(): "CMD_APO_START",
		reflect.TypeOf(&StartCookCommandV2{}).Elem(): "CMD_APO_START",
		reflect.TypeOf(&StopCookCommand{}).Elem():    "CMD_APO_STOP",

		reflect.TypeOf(&StartStageCommand{}).Elem():       "CMD_APO_START_STAGE",
		reflect.TypeOf(&UpdateCookStageCommand{}).Elem():  "CMD_APO_UPDATE_COOK_STAGE",
		reflect.TypeOf(&UpdateCookStagesCommand{}).Elem(): "CMD_APO_UPDATE_COOK_STAGES",

		reflect.TypeOf(&StartDescaleCommand{}).Elem(): "CMD_APO_START_DESCALE",
		reflect.TypeOf(&AbortDescaleCommand{}).Elem(): "CMD_APO_ABORT_DESCALE",

		reflect.TypeOf(&StartLiveStreamCommand{}).Elem(): "CMD_APO_START_LIVE_STREAM",
		reflect.TypeOf(&StopLiveStreamCommand{}).Elem():  "CMD_APO_STOP_LIVE_STREAM",

		reflect.TypeOf(&GenerateNewPairingCode{}).Elem(): "CMD_GENERATE_NEW_PAIRING",
		reflect.TypeOf(&AddUserWithPairingCode{}).Elem(): "CMD_ADD_USER_WITH_PAIRING",
		reflect.TypeOf(&ListUsersForDevice{}).Elem():     "CMD_LIST_USERS",
	}
)

type Command struct {
	ID      CookerID    `json:"id"`
	Type    MessageType `json:"type"`
	Payload interface{} `json:"payload"`
}

type SetFanCommand struct {
	Speed int `json:"speed"`
}

type HeatElementSetting struct {
	On bool `json:"on"`
}

type SetHeatingElementsCommand struct {
	Top    HeatingElementSetting `json:"top"`
	Rear   HeatingElementSetting `json:"rear"`
	Bottom HeatingElementSetting `json:"bottom"`
}

type SetLampCommand struct {
	On bool `json:"on"`
}

type SetLampPreferenceCommand struct {
	On bool `json:"on"`
}

type SetProbeCommand struct {
	Setpoint Temperature `json:"setpoint"`
}

type SetSteamGeneratorsCommand struct {
	Mode SteamGeneratorMode `json:"mode"`

	RelativeHumidity *SteamSetting `json:"relativeHumidity,omitempty"`
	SteamPercentage  *SteamSetting `json:"steamPercentage,omitempty"`
}

type SetTemperatureBulbsCommand struct {
	//TODO
}

type SetTemperatureUnitCommand struct {
	TemperatureUnit TemperatureUnit `json:"temperatureUnit"`
}

type SetTimerCommand struct {
	Initial int `json:"initial"`
}

type SetVentCommand struct {
	Open bool `json:"open"`
}

type OvenStage interface {
	isOvenStage()
}

type StageTemperatureProbe TemperatureSetting

// TODO merge with the Stage object in responses
type CookingStage struct {
	StepType string `json:"stepType"`

	ID                 string    `json:"id"`
	Type               StageType `json:"type"`
	UserActionRequired bool      `json:"userActionRequired"`

	Fan              *StageFan              `json:"fan"`
	HeatingElements  *StageHeatingElements  `json:"heatingElements"`
	TemperatureBulbs *StageTemperatureBulbs `json:"temperatureBulbs"`
	Vent             *StageVent             `json:"vent"`

	//TODO
	Title       *Title  `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`

	RackPosition    *StageRackPosition    `json:"rackPosition,omitempty"`
	SteamGenerators *StageSteamGenerators `json:"steamGenerators,omitempty"`

	// The "timerAdded" field is not present in the command schema?
	TimerAdded       *bool                  `json:"timerAdded,omitempty"`
	Timer            *StageTimer            `json:"timer,omitempty"`
	ProbeAdded       *bool                  `json:"probeAdded,omitempty"`
	TemperatureProbe *StageTemperatureProbe `json:"temperatureProbe,omitempty"`

	PhotoUrl          *string `json:"photoUrl,omitempty"`
	VideoThumbnailUrl *string `json:"videoThumbnailUrl,omitempty"`
	VideoUrl          *string `json:"videoUrl,omitempty"`
}

func (stage CookingStage) isOvenStage() {}

type StopStage struct {
	ID                 string    `json:"id"`
	Type               StageType `json:"type"`
	UserActionRequired bool      `json:"userActionRequired"`

	Title       *Title      `json:"title,omitempty"`
	Description *string     `json:"description,omitempty"`
	Timer       *StageTimer `json:"timer,omitempty"`
}

func (stage StopStage) isOvenStage() {}

type StartCookCommandV1 struct {
	CookID string `json:"cookId,omitempty"`
	//TODO make this OvenStage like it should be
	Stages []CookingStage `json:"stages"`
}

type StartCookCommandV2 struct {
	//TODO
}

type StartDescaleCommand struct{}

type AbortDescaleCommand struct{}

type StartFirmwareUpdateCommand struct {
	DownloadLink string `json:"downloadLink"`
}

type StartStageCommand struct {
	StageID string `json:"stageId"`
}

type StopCookCommand struct{}

type UpdateCookStageCommand CookingStage

type UpdateCookStagesCommand struct {
	Stages []CookingStage `json:"stages"`
}

type SetReportStateRate struct {
	Cooking int `json:"cooking"`
	Idle    int `json:"idle"`
}

type SetReportStateRateDefault struct{}

type SetBoilerTime struct {
	Time int `json:"time"`
}

type SupportedAccessory struct {
	Items string `json:"items"`
}

type AuthTokenV2 struct {
	Platform             string               `json:"platform"`
	SupportedAccessories []SupportedAccessory `json:"supported_accessories"`
	Token                string               `json:"token"`
}

type HealthCheck struct{}

type SetMetadataCommand struct {
	Metadata map[string]interface{} `json:"metadata"`
}

type GetConfiguration struct{}

type SetConfiguration struct {
	Token     string `json:"token"`
	ExpiresAt string `json:"expiresAt"`
}

type NameWifiDeviceCommand struct {
	Name string `json:"name"`
}

type DisconnectCommand struct {
	UserId string `json:"user_id,omitempty"`
}

type RequestDiagnosticsCommand struct {
	Command string `json:"command"`
}

type RegisterPushTokenCommand struct {
	AppId    string `json:"appId"`
	Platform string `json:"platform"`
	Token    string `json:"token"`
}

type TimeZone struct {
	ID        string `json:"id"`
	Code      string `json:"code"`
	GMTOffset int    `json:"gmt_offset"`
}

// Yes, these fields are snake_case
type SetTimeZoneCommand struct {
	TimeZone TimeZone `json:"time_zone"`
}

type StartLiveStreamCommand struct {
	Srt interface{} `json:"srt"`
}

type StopLiveStreamCommand struct{}

type GenerateNewPairingCode struct{}

type AddUserWithPairingCode struct {
	// JWT returned by GenerateNewPairingCode
	Data string `json:"data"`
}

type ListUsersForDevice struct{}

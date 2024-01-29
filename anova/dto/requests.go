package dto

import "reflect"

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

		reflect.TypeOf(&StartCookCommand{}).Elem(): "CMD_APO_START",
		reflect.TypeOf(&StopCookCommand{}).Elem():  "CMD_APO_STOP",

		reflect.TypeOf(&StartStageCommand{}).Elem():       "CMD_APO_START_STAGE",
		reflect.TypeOf(&UpdateCookStageCommand{}).Elem():  "CMD_APO_UPDATE_COOK_STAGE",
		reflect.TypeOf(&UpdateCookStagesCommand{}).Elem(): "CMD_APO_UPDATE_COOK_STAGES",

		reflect.TypeOf(&StartDescaleCommand{}).Elem(): "CMD_APO_START_DESCALE",
		reflect.TypeOf(&AbortDescaleCommand{}).Elem(): "CMD_APO_ABORT_DESCALE",

		reflect.TypeOf(&StartLiveStreamCommand{}).Elem(): "CMD_APO_START_LIVE_STREAM",
		reflect.TypeOf(&StopLiveStreamCommand{}).Elem():  "CMD_APO_STOP_LIVE_STREAM",

		reflect.TypeOf(&GenerateNewPairingCode{}).Elem(): "CMD_GENERATE_NEW_PAIRING",
		reflect.TypeOf(&AddUserWithPairingCode{}).Elem(): "CMD_ADD_USER_WITH_PAIRING",
	}
)

type Command struct {
	ID      CookerID    `json:"id"`
	Type    MessageType `json:"type"`
	Payload interface{} `json:"payload"`
}

type SetFanCommand struct {
	//TODO
}

type SetHeatingElementsCommand struct {
	//TODO
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
	//TODO
}

type SetTemperatureBulbsCommand struct {
	//TODO
}

type SetTemperatureUnitCommand struct {
	//TODO
}

type SetTimerCommand struct {
	//TODO
}

type SetVentCommand struct {
	//TODO
}

type StartCookCommand struct {
	//TODO
}

type StartDescaleCommand struct {
	//TODO
}

type AbortDescaleCommand struct {
	//TODO
}

type StartFirmwareUpdateCommand struct {
	//TODO
}

type StartStageCommand struct {
	//TODO
}

type StopCookCommand struct {
	//TODO
}

type UpdateCookStageCommand struct {
	//TODO
}

type UpdateCookStagesCommand struct {
	//TODO
}

type SetReportStateRate struct {
	//TODO
}

type SetReportStateRateDefault struct {
	//TODO
}

type SetBoilerTime struct {
	Time int `json:"time"`
}

type AuthTokenV2 struct {
	//TODO
}

type HealthCheck struct {
	//TODO
}

type SetMetadataCommand struct {
	Metadata map[string]interface{} `json:"metadata"`
}

type GetConfiguration struct {
	//TODO
}

type SetConfiguration struct {
	//TODO
}

type NameWifiDeviceCommand struct {
	Name string `json:"name"`
}

type DisconnectCommand struct {
	UserId string `json:"user_id,omitempty"`
}

type RequestDiagnosticsCommand struct {
	//TODO
}

type RegisterPushTokenCommand struct {
	//TODO
}

type SetTimeZoneCommand struct {
	//TODO
}

type StartLiveStreamCommand struct {
	//TODO
}

type StopLiveStreamCommand struct {
	//TODO
}

type GenerateNewPairingCode struct{}

type AddUserWithPairingCode struct {
	// JWT returned by GenerateNewPairingCode
	Data string `json:"data"`
}

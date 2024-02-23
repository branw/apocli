package anova

import (
	dto2 "go-apo/pkg/anova/dto"
	"log/slog"
	"math"
	"time"
)

// Pulled from Android app
const (
	minTempDryBulbCelsius = 25
	maxTempDryBulbCelsius = 250

	minTempWetBulbCelsius = 25
	maxTempWetBulbCelsius = 100

	minTempProbeCelsius = 1
	maxTempProbeCelsius = 100

	minTempDryBulbBottomHeatingCelsius = 25
	maxTempDryBulbBottomHeatingCelsius = 180

	maxTimerSeconds = 359940 // 99 hours 59 minutes
)

type Validatable interface {
	Validate() error
}

type StageEndCondition interface {
	Validatable

	isStageEndCondition()
}

type TimerTrigger string

const (
	TimerTriggerImmediately   TimerTrigger = "immediately"
	TimerTriggerWhenPreheated TimerTrigger = "preheat"
	TimerTriggerManually      TimerTrigger = "manually"
)

func (trigger TimerTrigger) Validate() error {
	if trigger != TimerTriggerImmediately && trigger != TimerTriggerWhenPreheated && trigger != TimerTriggerManually {
		return ErrInvalidTimerTrigger{}
	}
	return nil
}

type Timer struct {
	DurationSeconds int
	Trigger         TimerTrigger
}

func NewTimer(duration time.Duration, trigger TimerTrigger) *Timer {
	return &Timer{
		DurationSeconds: int(math.Round(duration.Seconds())),
		Trigger:         trigger,
	}
}

func (timer *Timer) isStageEndCondition() {}

func (timer *Timer) Validate() error {
	if err := timer.Trigger.Validate(); err != nil {
		return err
	}
	// See StageAdapter.isValidTimerValue
	if timer.DurationSeconds < 0 || timer.DurationSeconds > maxTimerSeconds {
		return ErrInvalidTimerDuration{}
	}
	return nil
}

type Probe struct {
	TemperatureCelsius float64
}

func NewProbeCelsius(temperature float64) *Probe {
	return &Probe{TemperatureCelsius: temperature}
}

func NewProbeFahrenheit(temperature float64) *Probe {
	return NewProbeCelsius(dto2.FahrenheitToCelsius(temperature))
}

func (probe *Probe) isStageEndCondition() {}

func (probe *Probe) Validate() error {
	// See StageAdapter.isValidTargetProbeTemperatureValue
	if probe.TemperatureCelsius < minTempProbeCelsius || probe.TemperatureCelsius > maxTempProbeCelsius {
		return ErrInvalidProbeTemperature{}
	}
	return nil
}

type HeatingElements struct {
	Top    bool
	Rear   bool
	Bottom bool
}

func NewHeatingElements(topOn bool, rearOn bool, bottomOn bool) HeatingElements {
	return HeatingElements{
		Top:    topOn,
		Rear:   rearOn,
		Bottom: bottomOn,
	}
}

var (
	RearElementOnly = NewHeatingElements(false, true, false)
)

func (heatingElements HeatingElements) Validate() error {
	// Valid combos: top, rear, bottom, top + rear, top + bottom
	if (!heatingElements.Top && !heatingElements.Rear && !heatingElements.Bottom) ||
		(heatingElements.Rear && heatingElements.Bottom) {
		return ErrInvalidHeatingElementCombination{}
	}
	return nil
}

type RackPosition int

const (
	RackPositionLow    RackPosition = 1
	RackPositionMiddle RackPosition = 3
	RackPositionHigh   RackPosition = 5
)

func (position RackPosition) Validate() error {
	if position < RackPositionLow || position > RackPositionHigh {
		return ErrInvalidRackPosition{}
	}
	return nil
}

type FanSpeed int

func NewFanSpeed(fanSpeed int) FanSpeed {
	return FanSpeed(fanSpeed)
}

const (
	FanSpeedHigh   FanSpeed = 100
	FanSpeedMedium FanSpeed = 67
	FanSpeedLow    FanSpeed = 33
	FanSpeedOff    FanSpeed = 0
)

func (speed FanSpeed) Validate() error {
	if speed < FanSpeedOff || speed > FanSpeedHigh {
		return ErrInvalidFanSpeed{}
	}
	return nil
}

type SteamPercentage float64

func NewSteamPercentage(percentage float64) *SteamPercentage {
	steamPercentage := SteamPercentage(percentage)
	return &steamPercentage
}

var (
	NoSteam *SteamPercentage = nil
)

func (percentage SteamPercentage) Validate() error {
	if percentage < 0 || percentage > 100 {
		return ErrInvalidSteamPercentage{}
	}
	return nil
}

type TemperatureMode string

const (
	TemperatureModeDry = "dry"
	TemperatureModeWet = "wet"
)

func (mode TemperatureMode) Validate() error {
	if mode != TemperatureModeDry && mode != TemperatureModeWet {
		return ErrInvalidTemperatureMode{}
	}
	return nil
}

type TemperatureSetpoint struct {
	TemperatureCelsius float64
	Mode               TemperatureMode
}

func NewSetpoint(temperatureCelsius float64, mode TemperatureMode) TemperatureSetpoint {
	return TemperatureSetpoint{
		TemperatureCelsius: temperatureCelsius,
		Mode:               mode,
	}
}

func NewSousVideSetpointCelsius(temperatureCelsius float64) TemperatureSetpoint {
	return NewSetpoint(temperatureCelsius, TemperatureModeWet)
}

func NewNonSousVideSetpointCelsius(temperatureCelsius float64) TemperatureSetpoint {
	return NewSetpoint(temperatureCelsius, TemperatureModeDry)
}

func (setpoint TemperatureSetpoint) Validate() error {
	if err := setpoint.Mode.Validate(); err != nil {
		return err
	}
	if (setpoint.Mode == TemperatureModeDry && (setpoint.TemperatureCelsius < minTempDryBulbCelsius || setpoint.TemperatureCelsius > maxTempDryBulbCelsius)) ||
		(setpoint.Mode == TemperatureModeWet && (setpoint.TemperatureCelsius < minTempWetBulbCelsius || setpoint.TemperatureCelsius > maxTempWetBulbCelsius)) {
		return ErrInvalidTemperatureSetpoint{}
	}
	return nil
}

type CookStage struct {
	RackPosition        RackPosition
	FanSpeed            FanSpeed
	TemperatureSetpoint TemperatureSetpoint
	HeatingElements     HeatingElements
	SteamPercentage     *SteamPercentage
	Terminator          StageEndCondition

	//TODO support transition settings between stages (currently, we always do automatic transitions)

	preheatStageId string
	cookStageId    string
	cook           *Cook
}

func NewCookStage(rackPosition RackPosition, fanSpeed FanSpeed, temperatureSetpoint TemperatureSetpoint, heatingElements HeatingElements, steamPercentage *SteamPercentage, terminator StageEndCondition) *CookStage {
	return &CookStage{
		RackPosition:        rackPosition,
		FanSpeed:            fanSpeed,
		TemperatureSetpoint: temperatureSetpoint,
		HeatingElements:     heatingElements,
		SteamPercentage:     steamPercentage,
		Terminator:          terminator,

		// Generate IDs for both a preheat and cook stage, even if the preheat stage is
		// unnecessary
		preheatStageId: generateRandomCookUuid(),
		cookStageId:    generateRandomCookUuid(),
	}
}

func (stage *CookStage) Validate() error {
	// Validate individual fields
	if err := stage.RackPosition.Validate(); err != nil {
		return err
	}
	if err := stage.FanSpeed.Validate(); err != nil {
		return err
	}
	if err := stage.TemperatureSetpoint.Validate(); err != nil {
		return err
	}
	if err := stage.HeatingElements.Validate(); err != nil {
		return err
	}
	if stage.SteamPercentage != nil {
		if err := (*stage.SteamPercentage).Validate(); err != nil {
			return err
		}
	}
	if stage.Terminator != nil {
		if err := stage.Terminator.Validate(); err != nil {
			return err
		}
	}

	// Validate combinations of fields
	if stage.FanSpeed != FanSpeedHigh && (stage.HeatingElements.Rear || stage.SteamPercentage != nil) {
		return ErrInvalidFanSpeed{}
	}

	return nil
}

//TODO fix; doesn't seem to do anything currently
//func (stage *CookStage) Update() error {
//	if stage.cook == nil {
//		return ErrNotAssociatedWithCook{}
//	}
//	if stage.cook.oven == nil {
//		return ErrCookNotStarted{}
//	}
//	return stage.cook.oven.UpdateCookStage(stage)
//}

func boolToPtr(value bool) *bool {
	v := value
	return &v
}

var (
	timerAdded *bool = boolToPtr(true)
	probeAdded *bool = boolToPtr(true)
)

func stageRackPositionToPtr(value dto2.StageRackPosition) *dto2.StageRackPosition {
	v := value
	return &v
}

// toDto converts a stage into one or more DTO stages. This is necessary to support timers
func (stage *CookStage) toDto() []dto2.CookingStage {
	description := "desc"
	cookStage := dto2.CookingStage{
		StepType: "stage",
		ID:       stage.cookStageId,
		Type:     dto2.StageTypeCook,
		// Updated below
		UserActionRequired: false,

		Fan: &dto2.StageFan{Speed: int(stage.FanSpeed)},
		HeatingElements: &dto2.StageHeatingElements{
			Top:    dto2.HeatingElementSetting{On: stage.HeatingElements.Top},
			Rear:   dto2.HeatingElementSetting{On: stage.HeatingElements.Rear},
			Bottom: dto2.HeatingElementSetting{On: stage.HeatingElements.Bottom},
		},
		TemperatureBulbs: &dto2.StageTemperatureBulbs{
			Mode: dto2.TemperatureBulbsMode(stage.TemperatureSetpoint.Mode),
			// Rest populated below
		},
		//TODO when is this ever not false?
		Vent: &dto2.StageVent{Open: false},

		RackPosition: stageRackPositionToPtr(dto2.StageRackPosition(stage.RackPosition)),

		Title:       dto2.NewTitle("foo bar"),
		Description: &description,
	}

	setpointTemperature := dto2.NewTemperatureFromCelsius(stage.TemperatureSetpoint.TemperatureCelsius)
	switch stage.TemperatureSetpoint.Mode {
	case TemperatureModeDry:
		cookStage.TemperatureBulbs.Dry = &dto2.TemperatureSetting{Setpoint: setpointTemperature}

		if stage.SteamPercentage != nil {
			cookStage.SteamGenerators = &dto2.StageSteamGenerators{
				Mode: dto2.SteamGeneratorModeSteamPercentage,
				SteamPercentage: &dto2.SteamSetting{
					Setpoint: int(*stage.SteamPercentage),
				},
			}
		}
	case TemperatureModeWet:
		cookStage.TemperatureBulbs.Wet = &dto2.TemperatureSetting{Setpoint: setpointTemperature}

		if stage.SteamPercentage != nil {
			cookStage.SteamGenerators = &dto2.StageSteamGenerators{
				Mode: dto2.SteamGeneratorModeRelativeHumidity,
				RelativeHumidity: &dto2.SteamSetting{
					Setpoint: int(*stage.SteamPercentage),
				},
			}
		}
	default:
		slog.Error("unknown temperature setpoint mode",
			slog.String("mode", string(stage.TemperatureSetpoint.Mode)))
	}

	needsPreheat := true
	if stage.Terminator != nil {
		switch terminator := stage.Terminator.(type) {
		case *Timer:
			if terminator.Trigger == TimerTriggerImmediately {
				needsPreheat = false
			} else if terminator.Trigger == TimerTriggerManually {
				cookStage.UserActionRequired = true
			}

			cookStage.TimerAdded = timerAdded
			cookStage.Timer = &dto2.StageTimer{Initial: terminator.DurationSeconds}

		case *Probe:
			cookStage.ProbeAdded = probeAdded
			cookStage.TemperatureProbe = &dto2.StageTemperatureProbe{
				Setpoint: dto2.NewTemperatureFromCelsius(terminator.TemperatureCelsius),
			}
		}
	}

	if !needsPreheat {
		return []dto2.CookingStage{cookStage}
	}

	preheatStage := cookStage
	preheatStage.ID = stage.preheatStageId
	preheatStage.Type = dto2.StageTypePreheat
	preheatStage.UserActionRequired = false

	return []dto2.CookingStage{preheatStage, cookStage}
}

type CookStages []*CookStage

func (stages CookStages) ToDto() []dto2.CookingStage {
	dtoStages := make([]dto2.CookingStage, 0)
	for _, stage := range stages {
		dtoStages = append(dtoStages, stage.toDto()...)
	}
	return dtoStages
}

type ErrNotAssociatedWithCook struct{}

func (err ErrNotAssociatedWithCook) Error() string {
	return "cook stage is not associated with a cook"
}

type ErrInvalidTimerTrigger struct{}

func (err ErrInvalidTimerTrigger) Error() string {
	return "invalid timer trigger"
}

type ErrInvalidTimerDuration struct{}

func (err ErrInvalidTimerDuration) Error() string {
	return "invalid timer duration"
}

type ErrInvalidProbeTemperature struct{}

func (err ErrInvalidProbeTemperature) Error() string {
	return "invalid probe temperature"
}

type ErrInvalidHeatingElementCombination struct{}

func (err ErrInvalidHeatingElementCombination) Error() string {
	return "invalid heating element combination"
}

type ErrInvalidRackPosition struct{}

func (err ErrInvalidRackPosition) Error() string {
	return "invalid rack position"
}

type ErrInvalidFanSpeed struct{}

func (err ErrInvalidFanSpeed) Error() string {
	return "invalid fan speed"
}

type ErrInvalidSteamPercentage struct{}

func (err ErrInvalidSteamPercentage) Error() string {
	return "invalid steam percentage"
}

type ErrInvalidTemperatureMode struct{}

func (err ErrInvalidTemperatureMode) Error() string {
	return "invalid temperature mode"
}

type ErrInvalidTemperatureSetpoint struct{}

func (err ErrInvalidTemperatureSetpoint) Error() string {
	return "invalid mode and temperature setpoint combination"
}

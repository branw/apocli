package anova

import (
	"errors"
	"testing"
)

func TestProbe_Validate(t *testing.T) {
	for _, probe := range []Probe{
		{},
		{TemperatureCelsius: -1},
		{TemperatureCelsius: 101},
	} {
		err := probe.Validate()
		if !errors.Is(err, ErrInvalidProbeTemperature{}) {
			t.Fatalf("expected failure for invalid temperature %+v", probe)
		}
	}

	for _, probe := range []Probe{
		{TemperatureCelsius: 1},
		{TemperatureCelsius: 25},
		{TemperatureCelsius: 100},
	} {
		err := probe.Validate()
		if err != nil {
			t.Fatalf("expected success for valid temperature %+v", probe)
		}
	}
}

func TestHeatingElements_Validate(t *testing.T) {
	for _, elements := range []HeatingElements{
		{},
		{Rear: true, Bottom: true},
		{Top: true, Rear: true, Bottom: true},
	} {
		err := elements.Validate()
		if !errors.Is(err, ErrInvalidHeatingElementCombination{}) {
			t.Fatalf("expected failure for invalid combination %+v", elements)
		}
	}

	for _, elements := range []HeatingElements{
		{Top: true},
		{Rear: true},
		{Bottom: true},
		{Top: true, Rear: true},
		{Top: true, Bottom: true},
	} {
		err := elements.Validate()
		if err != nil {
			t.Fatalf("expected success for valid combination %+v", elements)
		}
	}
}

func TestRackPosition_Validate(t *testing.T) {
	for _, position := range []RackPosition{
		RackPosition(0),
		RackPosition(6),
	} {
		err := position.Validate()
		if !errors.Is(err, ErrInvalidRackPosition{}) {
			t.Fatalf("expected failure for invalid rack position %+v", position)
		}
	}

	for _, position := range []RackPosition{
		RackPosition(1),
		RackPosition(2),
		RackPosition(5),
	} {
		err := position.Validate()
		if err != nil {
			t.Fatalf("expected success for valid rack position %+v", position)
		}
	}
}

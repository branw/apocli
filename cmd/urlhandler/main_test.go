package main

import (
	"apocli/pkg/anova"
	"github.com/stretchr/testify/assert"
	"net/url"
	"testing"
	"time"
)

func TestStartCookParams(t *testing.T) {
	t.Run("", func(t *testing.T) {
		inputUrl := "apo://default/start-cook?mode=wet&temp=80f&steam=60&elements=rear&timer=40m&trigger=button"

		parsedUrl, err := url.Parse(inputUrl)
		assert.Equal(t, nil, err)

		rawParams, err := url.ParseQuery(parsedUrl.RawQuery)
		assert.Equal(t, nil, err)

		params, err := parseStartCookParams(rawParams)
		assert.Equal(t, nil, err)
		assert.Equal(t, anova.TemperatureSetpoint{
			TemperatureCelsius: 27,
			Mode:               anova.TemperatureModeWet,
		}, params.temp)
		assert.Equal(t, anova.HeatingElements{
			Rear: true,
		}, params.elements)
		assert.Equal(t, anova.SteamPercentage(60), *params.steam)
		assert.Equal(t, anova.FanSpeedHigh, params.fanSpeed)
		assert.Equal(t, anova.NewTimer(40*time.Minute, anova.TimerTriggerManually), params.terminator)
	})
}

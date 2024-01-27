package dto

import (
	_ "embed"
	"github.com/xeipuuv/gojsonschema"
)

// JSON Schema for "EVENT_APO_STATE" messages
//
//go:embed schemas/oven_state_schema.json
var ovenStateSchemaJson string

// JSON Schema for outbound command messages
//
//go:embed schemas/oven_command_schema.json
var ovenCommandSchemaJson string

// JSON Schema for outbound multi-user command messages
//
//go:embed schemas/multi_user_command_schema.json
var multiUserCommandSchemaJson string

var (
	OvenStateSchema, _        = gojsonschema.NewSchema(gojsonschema.NewStringLoader(ovenStateSchemaJson))
	OvenCommandSchema, _      = gojsonschema.NewSchema(gojsonschema.NewStringLoader(ovenCommandSchemaJson))
	MultiUserCommandSchema, _ = gojsonschema.NewSchema(gojsonschema.NewStringLoader(multiUserCommandSchemaJson))
)

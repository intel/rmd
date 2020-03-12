package testhelpers

// BoolSchema schema
var BoolSchema = `{
		"type": "boolean"
	}`

// PositiveInteger schema
var PositiveInteger = `{
	"type": ["integer", "string"],
	"pattern": "^[0-9]*$", "minimum": 1
}`

// NonNegativeInteger schema
var NonNegativeInteger = `{
	"type": ["integer", "string"],
	"pattern": "^[0-9]*$", "minimum": 0
}`

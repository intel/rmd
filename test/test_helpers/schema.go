package testhelpers

var BoolSchema string = `{
		"type": "boolean"
	}`

var PositiveInteger string = `{
	"type": ["integer", "string"],
	"pattern": "^[0-9]*$", "minimum": 1
}`

var NonNegativeInteger = `{
	"type": ["integer", "string"],
	"pattern": "^[0-9]*$", "minimum": 0
}`

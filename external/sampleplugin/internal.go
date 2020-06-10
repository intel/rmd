// This file contains some internal structures and functions used by ModuleInterface implementation
// or REST handlers implementation. It's content is not exported neither directly accessible from
// the outside.

package main

import (
	"errors"
	"strings"
)

type inputData struct {
	name  string
	value float64
}

func convertParamsToData(params map[string]interface{}) (inputData, error) {
	result := inputData{}
	if len(params) == 0 {
		return inputData{}, errors.New("No params given")
	}
	// check if mandatory 'value' param exists
	valIface, ok := params["value"]
	if !ok {
		return inputData{}, errors.New("Lack of 'value' input param")
	}
	// check type of param
	valFloat, ok := valIface.(float64)
	if !ok {
		return inputData{}, errors.New("Invalid type of 'value' input param")
	}
	result.value = valFloat

	// check optional 'name' param
	nameIface, ok := params["name"]
	if ok {
		nameString, ok := nameIface.(string)
		if !ok {
			// optional param defined but has incorrect type
			return inputData{}, errors.New("Invalid type of 'name' input param")
		}
		if strings.ContainsAny(nameString, " \t") {
			return inputData{}, errors.New("Invalid content of 'name' input param")
		}
		result.name = nameString
	}

	return result, nil
}

package util

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestSubtractStringSlice(t *testing.T) {
	slice := []string{"a", "b", "c"}
	s := []string{"a", "c"}

	newslice := SubtractStringSlice(slice, s)

	if len(newslice) != 1 {
		t.Errorf("New slice length should be 1")
	}
	if newslice[0] != "b" {
		t.Errorf("New slice should be [\"2\"]")
	}
}

func TestUnifyMapParamsTypes(t *testing.T) {
	type args struct {
		pluginParams map[string]interface{}
	}
	ratioF64 := 1.5
	var myPstateFloat64 = map[string]interface{}{
		"ratio": ratioF64,
	}

	var ratioI64 int64 = 1
	var myPstateInt64 = map[string]interface{}{
		"ratio": ratioI64,
	}

	var ratioF32 float32 = 1.5
	var myPstateFloat32 = map[string]interface{}{
		"ratio": ratioF32,
	}

	var ratioI32 int32 = 1
	var myPstateInt32 = map[string]interface{}{
		"ratio": ratioI32,
	}

	var ratioJSONNumberF64 json.Number = "1.5"
	var myPstateJSONNumberF64 = map[string]interface{}{
		"ratio": ratioJSONNumberF64,
	}

	var ratioJSONNumberI64 json.Number = "1"
	var myPstateJSONNumberI64 = map[string]interface{}{
		"ratio": ratioJSONNumberI64,
	}

	paramInt64Table := []int64{8}
	var myParamInt64Table = map[string]interface{}{
		"param": paramInt64Table,
	}

	paramInt32Table := []int32{8}
	var myParamInt32Table = map[string]interface{}{
		"param": paramInt32Table,
	}

	paramFloat64Table := []float64{8}
	var myParamFloat64Table = map[string]interface{}{
		"param": paramFloat64Table,
	}

	paramFloat32Table := []float32{8}
	var myParamFloat32Table = map[string]interface{}{
		"param": paramFloat32Table,
	}

	paramString := "myText"
	var myParamString = map[string]interface{}{
		"param": paramString,
	}

	paramBool := true
	var myParamBool = map[string]interface{}{
		"param": paramBool,
	}

	var myParamNil = map[string]interface{}{
		"param": nil,
	}

	tests := []struct {
		name    string
		args    args
		want    map[string]interface{}
		wantErr bool
	}{
		{"float64", args{pluginParams: myPstateFloat64}, myPstateFloat64, false},
		{"float32", args{pluginParams: myPstateFloat32}, myPstateFloat64, false},
		{"int32", args{pluginParams: myPstateInt32}, myPstateInt64, false},
		{"json.Number to float64", args{pluginParams: myPstateJSONNumberF64}, myPstateFloat64, false},
		{"json.Number to int64", args{pluginParams: myPstateJSONNumberI64}, myPstateInt64, false},
		{"[]int64", args{pluginParams: myParamInt64Table}, myParamInt64Table, false},
		{"[]int32", args{pluginParams: myParamInt32Table}, myParamInt64Table, false},
		{"[]float64", args{pluginParams: myParamFloat64Table}, myParamFloat64Table, false},
		{"[]float32", args{pluginParams: myParamFloat32Table}, myParamFloat64Table, false},
		{"string", args{pluginParams: myParamString}, myParamString, false},
		{"bool", args{pluginParams: myParamBool}, myParamBool, false},
		{"nil", args{pluginParams: myParamNil}, myParamNil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UnifyMapParamsTypes(tt.args.pluginParams)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnifyMapParamsTypes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UnifyMapParamsTypes() = %v, want %v", got, tt.want)
			}
		})
	}
}

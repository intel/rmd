package util

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func typeConversion(value string, ntype string) (reflect.Value, error) {
	if ntype == "string" {
		return reflect.ValueOf(value), nil
	} else if ntype == "int" {
		i, err := strconv.Atoi(value)
		return reflect.ValueOf(i), err
	} else if ntype == "int8" {
		i, err := strconv.ParseInt(value, 10, 64)
		return reflect.ValueOf(int8(i)), err
	} else if ntype == "int32" {
		i, err := strconv.ParseInt(value, 10, 64)
		return reflect.ValueOf(int64(i)), err
	} else if ntype == "int64" {
		i, err := strconv.ParseInt(value, 10, 64)
		return reflect.ValueOf(i), err
	} else if ntype == "float32" {
		i, err := strconv.ParseFloat(value, 64)
		return reflect.ValueOf(float32(i)), err
	} else if ntype == "float64" {
		i, err := strconv.ParseFloat(value, 64)
		return reflect.ValueOf(i), err
	}
	//else if ....... add other type
	return reflect.ValueOf(value), fmt.Errorf("unknow type" + ntype)
}

// SetField sets obj's 'name' field with proper type
func SetField(obj interface{}, name string, value interface{}) error {
	structValue := reflect.ValueOf(obj).Elem()
	structFieldValue := structValue.FieldByName(name)

	if !structFieldValue.IsValid() {
		return fmt.Errorf("No such field: %s in obj", name)
	}

	if !structFieldValue.CanSet() {
		return fmt.Errorf("Cannot set %s field value", name)
	}

	val := reflect.ValueOf(value)
	structFieldType := structFieldValue.Type()

	switch structFieldType.Name() {
	// Try to remove these logic to TypeConversion
	case "int":
		v := value.(string)
		vInt, err := strconv.Atoi(v)
		if err != nil {
			// add log
			return err
		}
		val = reflect.ValueOf(vInt)
	}

	structFieldValue.Set(val)
	return nil
}

// StringInSlice string in slice
// O(n), maybe key in map is O(1)
// Not sure a good way for golang to check a string in slice
func StringInSlice(val string, list []string) bool {
	for _, v := range list {
		if v == val {
			return true
		}
	}
	return false
}

func stringReverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// Only support string slice at present
func sliceReverse(s []string) []string {

	reversed := []string{}

	// reverse order
	// and append into new slice
	for i := range s {
		n := s[len(s)-1-i]
		//fmt.Println(n) -- sanity check
		reversed = append(reversed, n)
	}
	return reversed
}

func initBitMap(num int) []byte {
	bitMap := make([]byte, num, num)
	for i := range bitMap {
		bitMap[i] = "0"[0]
	}
	return bitMap
}

// X86 is little endian
func setBitMap(scope string, bitmap []byte) error {
	if strings.Contains(scope, "-") {
		scopes := strings.SplitN(scope, "-", 2)
		low, err := strconv.Atoi(scopes[0])
		if err != nil {
			return err
		}
		high, err := strconv.Atoi(scopes[1])
		if err != nil {
			return err
		}
		if low >= len(bitmap) || high >= len(bitmap) {
			return fmt.Errorf("Set bitmap out index")
		}
		for i := low; i <= high; i++ {
			bitmap[i] = "1"[0]
		}
	} else {
		i, err := strconv.Atoi(scope)
		if err != nil {
			return err
		}
		if i >= len(bitmap) {
			return fmt.Errorf("Set bitmap out index")
		}
		bitmap[i] = "1"[0]
	}
	return nil
}

// cpus := []string{0: "1-2", 2: "3-5", 1: "7-9", 4: "80-87"}
// cpuBitMap := InitBitMap(88)
// for i, v := range cpus {
//     setBitMap(v, cpuBitMap)
// }
// cpustr := delimiterByComma(string(cpuBitMap), 24)
func delimiterByComma(bitstr string, step ...int) string {
	s := 32
	if len(step) > 0 {
		s = step[0]
	}
	len := len(bitstr)
	mod := len % s
	bitByte := []byte(bitstr)
	for i := len - mod; i > 0; i = i - s {
		head := string(bitByte[:i])
		tail := string(bitByte[i:])
		str := strings.Join([]string{head, tail}, ",")
		bitByte = []byte(str)
	}
	return string(bitByte)
}

func binary2Hex(str string) (string, error) {
	v, err := strconv.ParseInt(stringReverse(str), 2, 64)
	if err != nil {
		return "", err
	}
	return strconv.FormatUint(uint64(v), 16), err
}

// GenerateBitMap generate bit map
func GenerateBitMap(coreids []string, num int) (string, error) {
	cpuBitMap := initBitMap(num)
	for _, v := range coreids {
		err := setBitMap(v, cpuBitMap)
		if err != nil {
			return "", err
		}
	}

	cpus := delimiterByComma(stringReverse(string(cpuBitMap)))
	cpulist := strings.Split(cpus, ",")
	for i, str := range cpulist {
		v, _ := binary2Hex(str)
		cpulist[i] = v
	}

	return strings.Join(sliceReverse(cpulist), ","), nil
}

// IsZeroHexString is zero hex string
// No varification for input s.
// Such as whitespace in "0x11, 22" and "g" in "0xg1111"
// Caller do varification.
func IsZeroHexString(s string) bool {
	s = strings.TrimPrefix(strings.TrimPrefix(s, "0x"), "0X")
	s = strings.Replace(s, ",", "", -1)
	s = strings.TrimSpace(s)
	return len(s) == strings.Count(s, "0")
}

// HexMap cbm are all consecutive bits
var HexMap = map[byte]int{
	'1': 1,
	'3': 2,
	'7': 3,
	'f': 4,
	'F': 4,
}

// CbmLen returns length of CBM
func CbmLen(cbm string) int {
	len := 0
	for i := range cbm {
		len += HexMap[cbm[i]]
	}
	return len
}

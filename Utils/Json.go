package NoQ_RoomQ_Utils

import (
	"fmt"
	"reflect"
	"strconv"
)

type JSON struct {
	Key        string
	Val        interface{}
	Raw        string
	StatusCode int
}

func (j JSON) HasKey(key string) bool {
	_, ok := j.Val.(map[string]interface{})[key]
	return ok
}

func (j JSON) Get(key string) JSON {
	if val, ok := j.Val.(map[string]interface{}); ok {
		return JSON{Key: key, Val: val[key]}
	} else {
		panic("key not found " + key)
	}
}

func (j JSON) Index(index int) JSON {
	return JSON{Val: j.Val.([]interface{})[index]}
}

func (j JSON) String() string {
	return reflect.ValueOf(j.Val).String()
}

func (j JSON) Int() int64 {
	if v, err := strconv.ParseInt(fmt.Sprint(j.Val), 10, 64); err == nil {
		return v
	} else {
		fmt.Println(j.Raw)
		panic("can't parse value to int, Key: " + j.Key)
	}
}

func (j JSON) Uint() uint64 {
	if v, err := strconv.ParseUint(fmt.Sprint(j.Val), 10, 64); err == nil {
		return v
	} else {
		panic("can't parse value to uint, Key: " + j.Key)
	}
}

func (j JSON) Float() float64 {
	if v, err := strconv.ParseFloat(fmt.Sprint(j.Val), 64); err == nil {
		return v
	} else {
		panic("can't parse value to float, Key: " + j.Key)
	}
}

func (j JSON) Bool() bool {
	if v, err := strconv.ParseBool(fmt.Sprint(j.Val)); err == nil {
		return v
	} else {
		panic("can't parse value to bool, Key: " + j.Key)
	}
}

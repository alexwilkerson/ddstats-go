package main

import (
	"unicode/utf8"
)

type address uintptr

type gameVariable struct {
	parentOffset address
	offsets      []address
	variable     interface{}
}

func (gv *gameVariable) Get() {
	var pointer address
	getValue(&pointer, exeBaseAddress+gv.parentOffset)
	for i := 0; i < len(gv.offsets)-1; i++ {
		getValue(&pointer, pointer+gv.offsets[i])
	}
	getValue(&gv.variable, pointer+gv.offsets[len(gv.offsets)-1])
	switch gv.variable.(type) {
	case int:
		gv.variable = int(gv.variable.(int))
	case float32:
		gv.variable = float32(gv.variable.(float32))
	case float64:
		gv.variable = float64(gv.variable.(float64))
	case bool:
		gv.variable = bool(gv.variable.(bool))
	case string:
		gv.variable = string(gv.variable.(string))
	case address:
		gv.variable = address(gv.variable.(address))
	case uintptr:
		gv.variable = uintptr(gv.variable.(uintptr))
	}
}

func (gv *gameVariable) GetVariable() interface{} {
	return gv.variable
}

type gameStringVariable struct {
	lengthVariable gameVariable
	stringVariable gameVariable
	variable       string
}

func (gsv *gameStringVariable) Get() {
	gsv.lengthVariable.Get()
	size := gsv.lengthVariable.variable
	gsv.stringVariable.variable = string(make([]byte, size.(int)))
	gsv.stringVariable.Get()
	for !utf8.Valid([]byte(gsv.stringVariable.variable.(string))) {
		gsv.stringVariable.variable = string(make([]byte, size.(int)))
		gsv.stringVariable.offsets = append(gsv.stringVariable.offsets, 0x0)
		gsv.stringVariable.Get()
	}
	gsv.variable = gsv.stringVariable.variable.(string)[:size.(int)]
	gsv.lengthVariable.variable = 0
	gsv.stringVariable.variable = ""
}

func (gsv *gameStringVariable) GetVariable() interface{} {
	return string(gsv.variable)
}

type gameReplayIDVariable struct {
	replayIDVariable gameVariable
	variable         int
}

func (gridv *gameReplayIDVariable) GetVariable() interface{} {
	return int(gridv.variable)
}

func (gridv *gameReplayIDVariable) Get() {
	gridv.replayIDVariable.Get()
	gridv.variable = gridv.replayIDVariable.variable.(int)
	gridv.replayIDVariable.variable = "XXXXXX"
}

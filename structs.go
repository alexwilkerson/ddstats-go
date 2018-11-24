package main

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
	stringVariable gameVariable
	variable       string
}

// maxSize is used to check the maximum size of the string array in dd.
// if maxSize is 15, the pointer points directly to the char array
// if maxSize is 31 the char array holds the address of where the new
// string is stored. the offset of the maxSize is always 0x14.
func (gsv *gameStringVariable) Get() {
	lengthOffset := gsv.stringVariable.offsets[0] + 0x10
	lengthVariable := gameVariable{parentOffset: gameStatsAddress, offsets: []address{lengthOffset}, variable: 0}
	lengthVariable.Get()
	length := lengthVariable.variable

	maxSizeOffset := gsv.stringVariable.offsets[0] + 0x14
	maxSizeVariable := gameVariable{parentOffset: gameStatsAddress, offsets: []address{maxSizeOffset}, variable: 0}
	maxSizeVariable.Get()
	maxSize := maxSizeVariable.variable.(int)

	iterations := ((maxSize + 1) / 16) - 1

	for i := 0; i < iterations; i++ {
		gsv.stringVariable.offsets = append(gsv.stringVariable.offsets, 0x0)
	}
	gsv.stringVariable.variable = string(make([]byte, length.(int)))
	gsv.stringVariable.Get()

	gsv.variable = gsv.stringVariable.variable.(string)[:length.(int)]
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

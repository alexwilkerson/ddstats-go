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
	gsv.variable = gsv.stringVariable.variable.(string)
}

func (gsv *gameStringVariable) GetVariable() interface{} {
	return gsv.variable
}

type gameReplayIDVariable struct {
	replayIDVariable gameVariable
	variable         int
}

func (gridv *gameReplayIDVariable) GetVariable() interface{} {
	return gridv.variable
}

func (gridv *gameReplayIDVariable) Get() {
	gridv.replayIDVariable.Get()
	gridv.variable = gridv.replayIDVariable.variable.(int)
	gridv.replayIDVariable.variable = "XXXXXX"
}

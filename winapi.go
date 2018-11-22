package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"math"
	"reflect"
	"syscall"
	"unsafe"

	"github.com/TheTitanrain/w32"
)

func getHandle() (w32.HANDLE, error) {
	hwnd := w32.FindWindowW(nil, syscall.StringToUTF16Ptr("Devil Daggers"))
	if hwnd == 0 {
		return 0, errors.New("could not find Devil Daggers")
	}

	_, pid := w32.GetWindowThreadProcessId(hwnd)

	handle, err := w32.OpenProcess(w32.PROCESS_ALL_ACCESS, false, uintptr(pid))
	if err != nil {
		return 0, errors.New("could not open process Devil Daggers")
	}

	exeBaseAddress = getModuleBaseAddress(pid)

	return handle, nil
}

func getModuleBaseAddress(pid int) address {
	var baseAddress uintptr

	snapshot := w32.CreateToolhelp32Snapshot(w32.TH32CS_SNAPMODULE|w32.TH32CS_SNAPMODULE32, uint32(pid))
	if snapshot != w32.ERROR_INVALID_HANDLE {
		var me w32.MODULEENTRY32
		me.Size = uint32(unsafe.Sizeof(me))
		if w32.Module32First(snapshot, &me) {
			baseAddress = uintptr(unsafe.Pointer(me.ModBaseAddr))
		}
	}
	defer w32.CloseHandle(snapshot)

	return address(baseAddress)
}

func getAddressFrom(p address) address {
	buf, _, ok := w32.ReadProcessMemory(handle, uintptr(p), 4)
	if !ok {
		log.Fatalf("Error getting address from 0x%x.\n", p)
	}
	return toAddress(buf)
}

func getValue(i interface{}, p address) {
	if reflect.TypeOf(i).String() == "*main.address" {
		buf, _, ok := w32.ReadProcessMemory(handle, uintptr(p), 4)
		if !ok {
			log.Fatalf("Error getting address from 0x%x.\n", p)
		}
		vbuf := reflect.ValueOf(toAddress(buf))
		reflect.ValueOf(i).Elem().Set(vbuf)
		return
	}
	switch reflect.Indirect(reflect.ValueOf(i)).Elem().Type().String() {
	case "int":
		buf, _, ok := w32.ReadProcessMemory(handle, uintptr(p), 4)
		if !ok {
			log.Fatalf("Error getting int from 0x%x.\n", p)
		}
		vbuf := reflect.ValueOf(toInt(buf))
		reflect.ValueOf(i).Elem().Set(vbuf)
	case "float64":
		buf, _, ok := w32.ReadProcessMemory(handle, uintptr(p), 4)
		if !ok {
			log.Fatalf("Error getting int from 0x%x.\n", p)
		}
		vbuf := reflect.ValueOf(toFloat32(buf))
		reflect.ValueOf(i).Elem().Set(vbuf)
	case "bool":
		buf, _, ok := w32.ReadProcessMemory(handle, uintptr(p), 1)
		if !ok {
			log.Fatalf("Error getting int from 0x%x.\n", p)
		}
		vbuf := reflect.ValueOf(toBool(buf))
		reflect.ValueOf(i).Elem().Set(vbuf)
	case "string":
		sz := uintptr(len(reflect.Indirect(reflect.ValueOf(i)).Elem().String()))
		buf, _, ok := w32.ReadProcessMemory(handle, uintptr(p), sz)
		if !ok {
			log.Fatalf("Error getting int from 0x%x.\n", p)
		}
		vbuf := reflect.ValueOf(toString(buf))
		if reflect.Indirect(reflect.ValueOf(i)).Elem().String() == "XXXXXX" {
			reflect.ValueOf(i).Elem().Set(vbuf)
			var id int
			fmt.Sscanln(fmt.Sprint(reflect.ValueOf(vbuf)), &id)
			reflect.ValueOf(i).Elem().Set(reflect.ValueOf(id))
		} else {
			reflect.ValueOf(i).Elem().Set(vbuf)
		}
	}
}

func toString(b []uint16) string {
	var str []byte
	for _, c := range b {
		str = append(str, byte(int(c&0x00FF)))
		str = append(str, byte(int(c&0xFF00)>>8))
	}
	return string(str)
}

func toBool(b []uint16) bool {
	return int(b[0]&0x1) != 0
}

func toInt(b []uint16) int {
	return int(b[0]) | (int(b[1]) << 16) | (int(b[2]) << 32) | (int(b[3]) << 48)
}

func toFloat32(b []uint16) float32 {
	var flt []byte
	for _, c := range b {
		flt = append(flt, byte(int(c&0x00FF)))
		flt = append(flt, byte(int(c&0xFF00)>>8))
	}
	bits := binary.LittleEndian.Uint32(flt)
	float := math.Float32frombits(bits)
	return float
}

func toAddress(b []uint16) address {
	ret := address(0)
	for i := len(b) - 1; i >= 0; i-- {
		ret = (ret << 16) | address(b[i])
	}
	return ret
}

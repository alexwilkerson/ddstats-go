package main

import (
	"encoding/binary"
	"fmt"
	"math"
	"reflect"
	"syscall"
	"unsafe"

	"github.com/TheTitanrain/w32"
)

func getHandle() {
	var err error
	hwnd := w32.FindWindowW(nil, syscall.StringToUTF16Ptr("Devil Daggers"))
	if hwnd == 0 {
		handle = 0
		attached = false
		return
	}

	_, pid := w32.GetWindowThreadProcessId(hwnd)

	handle, err = w32.OpenProcess(w32.PROCESS_ALL_ACCESS, false, uintptr(pid))
	if err != nil {
		handle = 0
		attached = false
		return
	}

	exeBaseAddress, exeFilePath = getModuleInfo(pid)
	if len(exeFilePath) > 4 {
		survivalFilePath = exeFilePath[0:len(exeFilePath)-4] + "\\survival"
	}
	attached = true
}

func getModuleInfo(pid int) (address, string) {
	var baseAddress uintptr
	var exePath string

	snapshot := w32.CreateToolhelp32Snapshot(w32.TH32CS_SNAPMODULE|w32.TH32CS_SNAPMODULE32, uint32(pid))
	if snapshot != w32.ERROR_INVALID_HANDLE {
		var me w32.MODULEENTRY32
		me.Size = uint32(unsafe.Sizeof(me))
		if w32.Module32First(snapshot, &me) {
			baseAddress = uintptr(unsafe.Pointer(me.ModBaseAddr))
			exePath = syscall.UTF16ToString(me.SzExePath[:])
		}
	}
	defer w32.CloseHandle(snapshot)

	return address(baseAddress), exePath
}

func getAddressFrom(p address) address {
	buf, _, ok := w32.ReadProcessMemory(handle, uintptr(p), 4)
	if !ok {
		return p
		//log.Fatalf("Error getting address from 0x%x.\n", p)
	}
	return toAddress(buf)
}

func getValue(i interface{}, p address) {
	if reflect.TypeOf(i).String() == "*main.address" {
		buf, _, ok := w32.ReadProcessMemory(handle, uintptr(p), 4)
		if !ok {
			return
			// log.Fatalf("Error getting address from 0x%x.\n", p)
		}
		vbuf := reflect.ValueOf(toAddress(buf))
		reflect.ValueOf(i).Elem().Set(vbuf)
		return
	}
	switch reflect.Indirect(reflect.ValueOf(i)).Elem().Type().String() {
	case "int", "int32":
		buf, _, ok := w32.ReadProcessMemory(handle, uintptr(p), 4)
		if !ok {
			return
			// log.Fatalf("Error getting int from 0x%x.\n", p)
		}
		vbuf := reflect.ValueOf(toInt(buf))
		reflect.ValueOf(i).Elem().Set(vbuf)
	case "float32", "float64":
		buf, _, ok := w32.ReadProcessMemory(handle, uintptr(p), 4)
		if !ok {
			return
			// log.Fatalf("Error getting int from 0x%x.\n", p)
		}
		vbuf := reflect.ValueOf(toFloat32(buf))
		reflect.ValueOf(i).Elem().Set(vbuf)
	case "bool":
		buf, _, ok := w32.ReadProcessMemory(handle, uintptr(p), 1)
		if !ok {
			return
			// log.Fatalf("Error getting int from 0x%x.\n", p)
		}
		vbuf := reflect.ValueOf(toBool(buf))
		reflect.ValueOf(i).Elem().Set(vbuf)
	case "string":
		sz := uintptr(len(reflect.Indirect(reflect.ValueOf(i)).Elem().String()))
		buf, _, ok := w32.ReadProcessMemory(handle, uintptr(p), sz)
		if !ok {
			return
			// log.Fatalf("Error getting int from 0x%x.\n", p)
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
	return int(b[0]&0x0F) != 0
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

func setConsoleTitle(title string) (int, error) {
	handle, err := syscall.LoadLibrary("Kernel32.dll")
	if err != nil {
		return 0, err
	}
	defer syscall.FreeLibrary(handle)
	proc, err := syscall.GetProcAddress(handle, "SetConsoleTitleW")
	if err != nil {
		return 0, err
	}
	r, _, err := syscall.Syscall(proc, 1, uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(title))), 0, 0)
	return int(r), err
}

package winapi

import (
	"errors"
	"fmt"
	"syscall"
	"unsafe"

	"github.com/TheTitanrain/w32"
)

const (
	windowName = "Devil Daggers"
	// baseOffset should be updated if Devil Daggers is ever updated.
	baseOffset              = 0x00252760
	ddstatsBlockStartOffset = 0xEF4
)

// pointerOffsets should be updated if Devil Daggers is ever updated.
var pointerOffsets = []address{0x0, 0x30, 0x8, 0x60, 0x1A8}

type (
	handle  w32.HANDLE
	address uintptr
)

type WinAPI struct {
	connected           bool
	handle              handle
	baseAddress         address
	ddstatsBlockAddress address
	devilDaggersData    *DevilDaggersData
}

func New() *WinAPI {
	return &WinAPI{devilDaggersData: &DevilDaggersData{}}
}

// Connect makes a connection to the window with the given window name.
func (wa *WinAPI) Connect() error {
	hwnd := w32.FindWindowW(nil, syscall.StringToUTF16Ptr(windowName))
	if hwnd == 0 {
		wa.connected = false
		return fmt.Errorf("Connect: could not find window with name %q", windowName)
	}

	_, pid := w32.GetWindowThreadProcessId(hwnd)

	hndl, err := w32.OpenProcess(w32.PROCESS_ALL_ACCESS, false, uintptr(pid))
	if err != nil {
		wa.connected = false
		return fmt.Errorf("Connect: could not open process with name %q: %w", windowName, err)
	}

	baseAddress, err := getBaseAddress(pid)
	if err != nil {
		wa.connected = false
		return fmt.Errorf("Connect: could get base address: %w", err)
	}

	wa.connected = true
	wa.handle = handle(hndl)
	wa.baseAddress = baseAddress

	ddstatsBlockAddress, err := wa.getDDStatsBlockBaseAddress()
	if err != nil {
		wa.connected = false
		return fmt.Errorf("Connect: could get ddstats block address: %w", err)
	}

	wa.ddstatsBlockAddress = ddstatsBlockAddress

	return nil
}

func getBaseAddress(pid int) (address, error) {
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

	if baseAddress == 0 {
		return 0, fmt.Errorf("getBaseAddress: could not find base address for PID %d", pid)
	}

	return address(baseAddress), nil
}

func (wa *WinAPI) getDDStatsBlockBaseAddress() (address, error) {
	if wa.connected != true {
		return 0, errors.New("getAddressFromPointer: connection to window lost")
	}

	pointer, err := wa.getAddressFromPointer(wa.baseAddress + baseOffset)
	if err != nil {
		return 0, errors.New("getDDStatsBlockBaseAddress: could not get base pointer")
	}
	for i := range pointerOffsets {
		pointer, err = wa.getAddressFromPointer(pointer + pointerOffsets[i])
		if err != nil {
			return 0, errors.New("getDDStatsBlockBaseAddress: could not get base pointer")
		}
	}

	return pointer + ddstatsBlockStartOffset, nil
}

func (wa *WinAPI) getAddressFromPointer(p address) (address, error) {
	if wa.connected != true {
		return 0, errors.New("getAddressFromPointer: connection to window lost")
	}

	buf, _, ok := w32.ReadProcessMemory(w32.HANDLE(wa.handle), uintptr(p), 8)
	if !ok {
		return 0, errors.New("GetAddressFromPointer: unable to read process memory")
	}
	return toAddress(buf), nil
}

func toAddress(b []uint16) address {
	ret := address(0)
	for i := len(b) - 1; i >= 0; i-- {
		ret = (ret << 16) | address(b[i])
	}
	return ret
}

package ddstats

import (
	"errors"
	"fmt"
	"syscall"
	"time"
	"unsafe"

	"github.com/TheTitanrain/w32"
)

const (
	windowName = "Devil Daggers"
	// baseOffset should be updated if Devil Daggers is ever updated.
	baseOffset = 0x00252760
	// ddstatsBlockStartOffset should be updated if Devil Daggers is ever updated.
	ddstatsBlockStartOffset = 0xEF4
	defaultTickRate         = time.Second / 36
)

// pointerOffsets should be updated if Devil Daggers is ever updated.
var pointerOffsets = []address{0x0, 0x30, 0x8, 0x60, 0x1A8}

type (
	handle  w32.HANDLE
	address uintptr
)

// WinAPI is used to connect to and read data from Devil Daggers.
type DDStats struct {
	connected           bool
	handle              handle
	baseAddress         address
	ddstatsBlockAddress address
	dataBlock           *dataBlock
	tickRate            time.Duration
	done                chan struct{}
}

// New creates a new DDStats struct to use.
func New() *DDStats {
	done := make(chan struct{})
	return &DDStats{
		dataBlock: &dataBlock{},
		done:      done,
	}
}

func (dd *DDStats) WithTickRate(tickRate time.Duration) *DDStats {
	dd.tickRate = tickRate
	return dd
}

func (dd *DDStats) StartCapture(connected chan<- bool) {
	for {
		select {
		case <-time.After(dd.tickRate):
			if !dd.connected {
				err := dd.Connect()
				if err != nil {
					continue
				}
			}
			dd.RefreshData()
		case <-dd.done:
			fmt.Println("finished")
			break
		}
	}
}

func (dd *DDStats) StopCapture() {
	dd.done <- struct{}{}
}

func (dd *DDStats) GetConnected() bool {
	return dd.connected
}

// Connect attempts to make a connection to the Devil Daggers process.
func (dd *DDStats) Connect() error {
	hwnd := w32.FindWindowW(nil, syscall.StringToUTF16Ptr(windowName))
	if hwnd == 0 {
		dd.connected = false
		return fmt.Errorf("Connect: could not find window with name %q", windowName)
	}

	_, pid := w32.GetWindowThreadProcessId(hwnd)

	hndl, err := w32.OpenProcess(w32.PROCESS_ALL_ACCESS, false, uintptr(pid))
	if err != nil {
		dd.connected = false
		return fmt.Errorf("Connect: could not open process with name %q: %w", windowName, err)
	}

	baseAddress, err := getBaseAddress(pid)
	if err != nil {
		dd.connected = false
		return fmt.Errorf("Connect: could get base address: %w", err)
	}

	dd.connected = true
	dd.handle = handle(hndl)
	dd.baseAddress = baseAddress

	ddstatsBlockAddress, err := dd.getDDStatsBlockBaseAddress()
	if err != nil {
		dd.connected = false
		return fmt.Errorf("Connect: could get ddstats block address: %w", err)
	}

	dd.ddstatsBlockAddress = ddstatsBlockAddress

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

func (dd *DDStats) getDDStatsBlockBaseAddress() (address, error) {
	if dd.connected != true {
		return 0, errors.New("getAddressFromPointer: connection to window lost")
	}

	pointer, err := dd.getAddressFromPointer(dd.baseAddress + baseOffset)
	if err != nil {
		return 0, errors.New("getDDStatsBlockBaseAddress: could not get base pointer")
	}
	for i := range pointerOffsets {
		pointer, err = dd.getAddressFromPointer(pointer + pointerOffsets[i])
		if err != nil {
			return 0, errors.New("getDDStatsBlockBaseAddress: could not get base pointer")
		}
	}

	return pointer + ddstatsBlockStartOffset, nil
}

func (dd *DDStats) getAddressFromPointer(p address) (address, error) {
	if dd.connected != true {
		return 0, errors.New("getAddressFromPointer: connection to window lost")
	}

	buf, _, ok := w32.ReadProcessMemory(w32.HANDLE(dd.handle), uintptr(p), 8)
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

// SetConsoleTitle sets the console title.
func (dd *DDStats) SetConsoleTitle(title string) error {
	handle, err := syscall.LoadLibrary("Kernel32.dll")
	if err != nil {
		return err
	}
	defer syscall.FreeLibrary(handle)
	proc, err := syscall.GetProcAddress(handle, "SetConsoleTitleW")
	if err != nil {
		return err
	}
	_, _, err = syscall.Syscall(proc, 1, uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(title))), 0, 0)
	return err
}
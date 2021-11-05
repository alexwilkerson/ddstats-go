package devildaggers

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
	// This is the offset for "dd.exe" + baseOffset found in Cheat Engine.
	baseOffset                   = 0x00250DC0
	persistentConnectionTickRate = time.Second / 60
)

const windowsCodeStillActive = 259

var deathTypes = []string{"Fallen", "Swarmed", "Impaled", "Gored", "Infested", "Opened", "Purged",
	"Desecrated", "Sacrificed", "Eviscerated", "Annihilated", "Intoxicated",
	"Envenmonated", "Incarnated", "Discarnated", "Barbed"}

type (
	handle  w32.HANDLE
	address uintptr
)

// DevilDaggers is used to connect to and read data from Devil Daggers.
type DevilDaggers struct {
	connected           bool
	handle              handle
	baseAddress         address
	ddstatsBlockAddress address
	dataBlock           *DataBlock
	statsFrame          []StatsFrame
	errors              chan error
	done                chan struct{}
}

// New creates a new DDStats struct to use.
func New() *DevilDaggers {
	return &DevilDaggers{
		dataBlock:  &DataBlock{},
		statsFrame: []StatsFrame{},
	}
}

func (dd *DevilDaggers) StartPersistentConnection(errors chan error) {
	if dd.done != nil {
		close(dd.done)
	}
	dd.done = make(chan struct{})
	go func() {
		for {
			select {
			case <-time.After(persistentConnectionTickRate):
				if !dd.connected {
					connected, err := dd.Connect()
					if err != nil {
						// errors <- fmt.Errorf("StartPersistentConnection: could not connect to devil daggers: %w", err)
						continue
					}
					dd.connected = connected
					if !dd.connected {
						continue
					}
				}
				dd.connected = dd.checkConnection()
				if dd.connected {
					err := dd.RefreshData()
					if err != nil {
						errors <- fmt.Errorf("StartPersistentConnection: could not refresh data: %w", err)
						continue
					}
				}
			case <-dd.done:
				dd.Close()
				dd.connected = false
				return
			}
		}
	}()
}

func (dd *DevilDaggers) StopPersistentConnection() {
	if dd.done != nil {
		close(dd.done)
	}
}

// Connect attempts to make a connection to the Devil Daggers process.
func (dd *DevilDaggers) Connect() (bool, error) {
	hwnd := w32.FindWindowW(nil, syscall.StringToUTF16Ptr(windowName))
	if hwnd == 0 {
		dd.connected = false
		return false, nil
	}

	_, pid := w32.GetWindowThreadProcessId(hwnd)

	hndl, err := w32.OpenProcess(w32.PROCESS_ALL_ACCESS, false, uintptr(pid))
	if err != nil {
		dd.connected = false
		return false, fmt.Errorf("Connect: could not open process with name %q: %w", windowName, err)
	}

	baseAddress, err := getBaseAddress(pid)
	if err != nil {
		dd.connected = false
		return false, fmt.Errorf("Connect: could get base address: %w", err)
	}

	dd.connected = true
	dd.handle = handle(hndl)
	dd.baseAddress = baseAddress

	ddstatsBlockAddress, err := dd.getDevilDaggersBlockBaseAddress()
	if err != nil {
		dd.connected = false
		return false, fmt.Errorf("Connect: could get ddstats block address: %w", err)
	}

	dd.ddstatsBlockAddress = ddstatsBlockAddress

	return true, nil
}

// Close closes the handle to Devil Daggers.
func (dd *DevilDaggers) Close() {
	w32.CloseHandle(w32.HANDLE(dd.handle))
}

func (dd *DevilDaggers) CheckConnection() bool {
	return dd.connected
}

// CheckConnection returns whether the DevilDaggers struct is currently connected to Devil Daggers.
func (dd *DevilDaggers) checkConnection() bool {
	code, err := w32.GetExitCodeProcess(w32.HANDLE(dd.handle))
	if err != nil || code != windowsCodeStillActive {
		return false
	}

	return true
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

func (dd *DevilDaggers) getDevilDaggersBlockBaseAddress() (address, error) {
	if dd.connected != true {
		return 0, errors.New("getAddressFromPointer: connection to window lost")
	}

	pointer, err := dd.getAddressFromPointer(dd.baseAddress + baseOffset)
	if err != nil {
		return 0, errors.New("getDevilDaggersBlockBaseAddress: could not get base pointer")
	}

	// 0xC skips the '__ddstats__' header in the memory block.
	return pointer + 0xC, nil
}

func (dd *DevilDaggers) getAddressFromPointer(p address) (address, error) {
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

func GetDeathTypeString(deathType int) (string, error) {
	if deathType < 0 || deathType > len(deathTypes) {
		return "", errors.New("GetDeathTypeString: no death type related to this int")
	}

	return deathTypes[deathType], nil
}

package winapi

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/TheTitanrain/w32"
)

type DevilDaggersData struct {
	DDStatsVersion     uint32
	PlayerID           uint32
	UserName           [32]byte
	Time               float32
	GemsCollected      uint32
	Kills              uint32
	DaggersFired       uint32
	DaggersHit         uint32
	EnemiesAlive       uint32
	LevelGems          uint32
	HomingDaggers      uint32
	GemsDespawns       uint32
	GemsEaten          uint32
	TotalGems          uint32
	PerEnemyAliveCount [17]uint16
	PerEnemyKillCount  [17]uint16
	IsPlayerAlive      bool
	IsReplay           bool
	DeathType          uint8
	IsInGame           bool
	ReplayPlayerID     uint32
	ReplayPlayerName   [32]byte
	LevelHashMD5       [16]byte
	TimeLvl2           float32
	TimeLvl3           float32
	TimeLvl4           float32
	LeviDownTime       float32
	OrbDownTime        float32
	Status             uint32
}

func (wa *WinAPI) RefreshDevilDaggersDataBlock() error {
	if wa.connected != true {
		return errors.New("RefreshDevilDaggersDataBlock: connection to window lost")
	}

	buf, _, ok := w32.ReadProcessMemory(w32.HANDLE(wa.handle), uintptr(wa.ddstatsBlockAddress), 232)
	if !ok {
		return errors.New("RefreshDevilDaggersDataBlock: unable to read process memory")
	}

	byteBuf := bytes.Buffer{}

	for _, b := range buf {
		split := make([]byte, 2)
		binary.LittleEndian.PutUint16(split, b)
		byteBuf.Write(split)
	}

	err := binary.Read(&byteBuf, binary.LittleEndian, wa.devilDaggersData)
	if err != nil {
		return fmt.Errorf("RefreshDevilDaggersDataBlock: unable to encode data block: %w", err)
	}

	fmt.Printf("%+v", wa.devilDaggersData)

	return nil
}

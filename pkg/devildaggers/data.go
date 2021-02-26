package devildaggers

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"unsafe"

	"github.com/TheTitanrain/w32"
)

const dataSize = 232

const (
	skull1 = iota
	skull2
	spiderling
	skull3
	squid1
	squid2
	squid3
	centipede1
	centipede2
	spider1
	spider2
	leviathan
	orb
	thorn
	centipede3
	spiderEgg
)

const (
	// StatusTitle is when the user is in the title screen.
	StatusTitle uint32 = iota
	// StatusMenu is when the user is in the menu.
	StatusMenu
	// StatusLobby is when the user is in the dagger lobby.
	StatusLobby
	// StatusPlaying is when the user is playing the game.
	StatusPlaying
	// StatusDead is when the user is dead.
	StatusDead
	// StatusOwnReplay is when the user is watching their own replay
	StatusOwnReplay
	// StatusOtherReplay is when the user is watching another person's replay.
	StatusOtherReplay
)

type dataBlock struct {
	DDStatsVersion       uint32
	PlayerID             uint32
	UserName             [32]byte
	Time                 float32
	GemsCollected        uint32
	Kills                uint32
	DaggersFired         uint32
	DaggersHit           uint32
	EnemiesAlive         uint32
	LevelGems            uint32
	HomingDaggers        uint32
	GemsDespawned        uint32
	GemsEaten            uint32
	TotalGems            uint32
	DaggersEaten         uint32
	PerEnemyAliveCount   [17]uint16
	PerEnemyKillCount    [17]uint16
	IsPlayerAlive        bool
	IsReplay             bool
	DeathType            uint8
	IsInGame             bool
	ReplayPlayerID       uint32
	ReplayPlayerName     [32]byte
	LevelHashMD5         [16]byte
	TimeLvl2             float32
	TimeLvl3             float32
	TimeLvl4             float32
	LeviDownTime         float32
	OrbDownTime          float32
	Status               uint32
	HomingMax            uint32
	TimeHomingMax        float32
	EnemiesAliveMax      uint32
	TimeEnemiesAliveMax  float32
	TimeMax              float32
	Padding1             [4]byte // cause computer science
	StatsBase            uint64  // address to stat frame array
	StatsFramesLoaded    uint32
	StatsFinishedLoading bool
	Padding2             [3]byte // Padding here because previous data is in a struct with a pointer.
	StartingHandLevel    uint32
	StartingHomingCount  uint32
	StartingTime         float32
	ProhibitedMods       bool
}

type statsFrame struct {
	GemsCollected      uint32
	Kills              uint32
	DaggersFired       uint32
	DaggersHit         uint32
	EnemiesAlive       uint32
	LevelGems          uint32
	HomingDaggers      uint32
	GemsDespawned      uint32
	GemsEaten          uint32
	TotalGems          uint32
	PerEnemyAliveCount [17]uint16
	PerEnemyKillCount  [17]uint16
}

// RefreshData attempts to read the Devil Daggers process memory. The data is acquired based
// on the __ddstats__ block within the game's memory. The data is then read into the dataBlock
// struct. The variables of this data can then be read using the various 'Get' methods.
func (dd *DevilDaggers) RefreshData() error {
	if dd.connected != true {
		return errors.New("RefreshData: connection to window lost")
	}

	buf, _, ok := w32.ReadProcessMemory(w32.HANDLE(dd.handle), uintptr(dd.ddstatsBlockAddress), unsafe.Sizeof(*dd.dataBlock))
	if !ok {
		return errors.New("RefreshData: unable to read process memory")
	}

	byteBuf := bytes.NewBuffer(make([]byte, 0, len(buf)*2))

	for _, b := range buf {
		split := make([]byte, 2)
		binary.LittleEndian.PutUint16(split, b)
		byteBuf.Write(split)
	}

	err := binary.Read(byteBuf, binary.LittleEndian, dd.dataBlock)
	if err != nil {
		fmt.Println(err)
		return fmt.Errorf("RefreshData: unable to encode data block: %w", err)
	}

	err = dd.RefreshStatsFrame()
	if err != nil {
		fmt.Println(err)
		return fmt.Errorf("RefreshData: unable to refresh stats frame: %w", err)
	}

	return nil
}

func (dd *DevilDaggers) RefreshStatsFrame() error {
	if dd.connected != true {
		return nil
	}

	framesLoaded := int(dd.dataBlock.StatsFramesLoaded)

	if cap(dd.statsFrame) < framesLoaded {
		dd.statsFrame = append(dd.statsFrame, make([]statsFrame, framesLoaded-cap(dd.statsFrame))...)
	}

	dd.statsFrame = dd.statsFrame[:framesLoaded]

	if framesLoaded == 0 {
		return nil
	}

	buf, _, ok := w32.ReadProcessMemory(w32.HANDLE(dd.handle), uintptr(dd.dataBlock.StatsBase), unsafe.Sizeof(dd.statsFrame[0])*uintptr(framesLoaded))
	if !ok {
		return errors.New("RefreshStatsFrame: unable to read process memory")
	}

	byteBuf := bytes.NewBuffer(make([]byte, 0, len(buf)*2))

	for _, b := range buf {
		split := make([]byte, 2)
		binary.LittleEndian.PutUint16(split, b)
		byteBuf.Write(split)
	}

	err := binary.Read(byteBuf, binary.LittleEndian, dd.statsFrame)
	if err != nil {
		fmt.Println(err)
		return fmt.Errorf("RefreshStatsFrame: unable to encode data block: %w", err)
	}

	fmt.Println(dd.statsFrame[1])

	return nil

}

func (dd *DevilDaggers) GetDDStatsVersion() uint32 {
	return dd.dataBlock.DDStatsVersion
}

func (dd *DevilDaggers) GetPlayerID() uint32 {
	return dd.dataBlock.PlayerID
}

func (dd *DevilDaggers) GetPlayerName() string {
	return byteArrayToString(&dd.dataBlock.UserName)
}

func (dd *DevilDaggers) GetTime() float32 {
	return dd.dataBlock.Time
}

func (dd *DevilDaggers) GetGemsCollected() uint32 {
	return dd.dataBlock.GemsCollected
}

func (dd *DevilDaggers) GetKills() uint32 {
	return dd.dataBlock.Kills
}

func (dd *DevilDaggers) GetDaggersFired() uint32 {
	return dd.dataBlock.DaggersFired
}

func (dd *DevilDaggers) GetDaggersHit() uint32 {
	return dd.dataBlock.DaggersHit
}

func (dd *DevilDaggers) GetAccuracy() float32 {
	if dd.dataBlock.DaggersFired == 0 {
		return 0.0
	}
	return float32(dd.dataBlock.DaggersHit) / float32(dd.dataBlock.DaggersFired) * 100
}

func (dd *DevilDaggers) GetEnemiesAlive() uint32 {
	return dd.dataBlock.EnemiesAlive
}

func (dd *DevilDaggers) GetLevelGems() uint32 {
	return dd.dataBlock.LevelGems
}

func (dd *DevilDaggers) GetHomingDaggers() uint32 {
	return dd.dataBlock.HomingDaggers
}

func (dd *DevilDaggers) GetGemsDespawned() uint32 {
	return dd.dataBlock.GemsDespawned
}

func (dd *DevilDaggers) GetGemsEaten() uint32 {
	return dd.dataBlock.GemsEaten
}

func (dd *DevilDaggers) GetTotalGems() uint32 {
	return dd.dataBlock.TotalGems
}

func (dd *DevilDaggers) GetDaggersEaten() uint32 {
	return dd.dataBlock.TotalGems
}

func (dd *DevilDaggers) GetSkull1Alive() uint16 {
	return dd.dataBlock.PerEnemyAliveCount[skull1]
}

func (dd *DevilDaggers) GetSkull2Alive() uint16 {
	return dd.dataBlock.PerEnemyAliveCount[skull2]
}

func (dd *DevilDaggers) GetSpiderlingAlive() uint16 {
	return dd.dataBlock.PerEnemyAliveCount[spiderling]
}

func (dd *DevilDaggers) GetSkull3Alive() uint16 {
	return dd.dataBlock.PerEnemyAliveCount[skull3]
}

func (dd *DevilDaggers) GetSquid1Alive() uint16 {
	return dd.dataBlock.PerEnemyAliveCount[squid1]
}

func (dd *DevilDaggers) GetSquid2Alive() uint16 {
	return dd.dataBlock.PerEnemyAliveCount[squid2]
}

func (dd *DevilDaggers) GetSquid3Alive() uint16 {
	return dd.dataBlock.PerEnemyAliveCount[squid3]
}

func (dd *DevilDaggers) GetCentipede1Alive() uint16 {
	return dd.dataBlock.PerEnemyAliveCount[centipede1]
}

func (dd *DevilDaggers) GetCentipede2Alive() uint16 {
	return dd.dataBlock.PerEnemyAliveCount[centipede2]
}

func (dd *DevilDaggers) GetSpider1Alive() uint16 {
	return dd.dataBlock.PerEnemyAliveCount[spider1]
}

func (dd *DevilDaggers) GetSpider2Alive() uint16 {
	return dd.dataBlock.PerEnemyAliveCount[spider2]
}

func (dd *DevilDaggers) GetLeviathanAlive() uint16 {
	return dd.dataBlock.PerEnemyAliveCount[leviathan]
}

func (dd *DevilDaggers) GetOrbAlive() uint16 {
	return dd.dataBlock.PerEnemyAliveCount[orb]
}

func (dd *DevilDaggers) GetThornAlive() uint16 {
	return dd.dataBlock.PerEnemyAliveCount[thorn]
}

func (dd *DevilDaggers) GetCentipede3Alive() uint16 {
	return dd.dataBlock.PerEnemyAliveCount[centipede3]
}

func (dd *DevilDaggers) GetSpiderEggAlive() uint16 {
	return dd.dataBlock.PerEnemyAliveCount[spiderEgg]
}

func (dd *DevilDaggers) GetSkull1Killed() uint16 {
	return dd.dataBlock.PerEnemyKillCount[skull1]
}

func (dd *DevilDaggers) GetSkull2Killed() uint16 {
	return dd.dataBlock.PerEnemyKillCount[skull2]
}

func (dd *DevilDaggers) GetSpiderlingKilled() uint16 {
	return dd.dataBlock.PerEnemyKillCount[spiderling]
}

func (dd *DevilDaggers) GetSkull3Killed() uint16 {
	return dd.dataBlock.PerEnemyKillCount[skull3]
}

func (dd *DevilDaggers) GetSquid1Killed() uint16 {
	return dd.dataBlock.PerEnemyKillCount[squid1]
}

func (dd *DevilDaggers) GetSquid2Killed() uint16 {
	return dd.dataBlock.PerEnemyKillCount[squid2]
}

func (dd *DevilDaggers) GetSquid3Killed() uint16 {
	return dd.dataBlock.PerEnemyKillCount[squid3]
}

func (dd *DevilDaggers) GetCentipede1Killed() uint16 {
	return dd.dataBlock.PerEnemyKillCount[centipede1]
}

func (dd *DevilDaggers) GetCentipede2Killed() uint16 {
	return dd.dataBlock.PerEnemyKillCount[centipede2]
}

func (dd *DevilDaggers) GetSpider1Killed() uint16 {
	return dd.dataBlock.PerEnemyKillCount[spider1]
}

func (dd *DevilDaggers) GetSpider2Killed() uint16 {
	return dd.dataBlock.PerEnemyKillCount[spider2]
}

func (dd *DevilDaggers) GetLeviathanKilled() uint16 {
	return dd.dataBlock.PerEnemyKillCount[leviathan]
}

func (dd *DevilDaggers) GetOrbKilled() uint16 {
	return dd.dataBlock.PerEnemyKillCount[orb]
}

func (dd *DevilDaggers) GetThornKilled() uint16 {
	return dd.dataBlock.PerEnemyKillCount[thorn]
}

func (dd *DevilDaggers) GetCentipede3Killed() uint16 {
	return dd.dataBlock.PerEnemyKillCount[centipede3]
}

func (dd *DevilDaggers) GetSpiderEggKilled() uint16 {
	return dd.dataBlock.PerEnemyKillCount[spiderEgg]
}

func (dd *DevilDaggers) GetIsPlayerAlive() bool {
	return dd.dataBlock.IsPlayerAlive
}

func (dd *DevilDaggers) GetIsReplay() bool {
	return dd.dataBlock.IsReplay
}

func (dd *DevilDaggers) GetDeathType() uint8 {
	return dd.dataBlock.DeathType
}

func (dd *DevilDaggers) GetIsInGame() bool {
	return dd.dataBlock.IsInGame
}

func (dd *DevilDaggers) GetReplayPlayerID() uint32 {
	return dd.dataBlock.ReplayPlayerID
}

func (dd *DevilDaggers) GetReplayPlayerName() string {
	return byteArrayToString(&dd.dataBlock.ReplayPlayerName)
}

func (dd *DevilDaggers) GetLevelHashMD5() string {
	return fmt.Sprintf("%X", dd.dataBlock.LevelHashMD5)
}

func (dd *DevilDaggers) GetTimeLvl2() float32 {
	return dd.dataBlock.TimeLvl2
}

func (dd *DevilDaggers) GetTimeLvl3() float32 {
	return dd.dataBlock.TimeLvl3
}

func (dd *DevilDaggers) GetTimeLvl4() float32 {
	return dd.dataBlock.TimeLvl4
}

func (dd *DevilDaggers) GetLeviathanDownTime() float32 {
	return dd.dataBlock.LeviDownTime
}

func (dd *DevilDaggers) GetOrbDownTime() float32 {
	return dd.dataBlock.OrbDownTime
}

func (dd *DevilDaggers) GetStatus() uint32 {
	return dd.dataBlock.Status
}

func (dd *DevilDaggers) GetHomingMax() uint32 {
	return dd.dataBlock.HomingMax
}

func (dd *DevilDaggers) GetHomingMaxTime() float32 {
	return dd.dataBlock.TimeHomingMax
}

func (dd *DevilDaggers) GetEnemiesAliveMax() uint32 {
	return dd.dataBlock.EnemiesAliveMax
}

func (dd *DevilDaggers) GetEnemiesAliveMaxTime() float32 {
	return dd.dataBlock.TimeEnemiesAliveMax
}

func (dd *DevilDaggers) GetTimeMax() float32 {
	return dd.dataBlock.TimeMax
}

func (dd *DevilDaggers) GetStatsFramesLoaded() uint32 {
	return dd.dataBlock.StatsFramesLoaded
}

func (dd *DevilDaggers) GetStatsFinishedLoading() bool {
	return dd.dataBlock.StatsFinishedLoading
}

func (dd *DevilDaggers) GetStartingHandLevel() uint32 {
	return dd.dataBlock.StartingHandLevel
}

func (dd *DevilDaggers) GetStartingHomingCount() uint32 {
	return dd.dataBlock.StartingHomingCount
}

func (dd *DevilDaggers) GetStartingTime() float32 {
	return dd.dataBlock.StartingTime
}

func (dd *DevilDaggers) GetProhibitedMods() bool {
	return dd.dataBlock.ProhibitedMods
}

func byteArrayToString(a *[32]byte) string {
	for i, b := range a {
		if b == 0 {
			return string(a[:i])
		}
	}
	return string(a[:])
}

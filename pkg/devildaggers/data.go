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
	StatusTitle int32 = iota
	// StatusMenu is when the user is in the menu.
	StatusMenu
	// StatusLobby is when the user is in the dagger lobby.
	StatusLobby
	// StatusPlaying is when the user is playing the game.
	StatusPlaying
	// StatusDead is when the user is dead.
	StatusDead
	// StatusOwnReplayFromLastRun is when the user is watching their own replay directly after a run.
	StatusOwnReplayFromLastRun
	// StatusOwnReplayFromLeaderboard is when the user is watching their own replay from the leaderboard.
	StatusOwnReplayFromLeaderboard
	// StatusOtherReplay is when the user is watching another person's replay.
	StatusOtherReplay
)

type DataBlock struct {
	DDStatsVersion       int32
	PlayerID             int32
	UserName             [32]byte
	Time                 float32
	GemsCollected        int32
	Kills                int32
	DaggersFired         int32
	DaggersHit           int32
	EnemiesAlive         int32
	LevelGems            int32
	HomingDaggers        int32
	GemsDespawned        int32
	GemsEaten            int32
	TotalGems            int32
	DaggersEaten         int32
	PerEnemyAliveCount   [17]int16
	PerEnemyKillCount    [17]int16
	IsPlayerAlive        bool
	IsReplay             bool
	DeathType            uint8
	IsInGame             bool
	ReplayPlayerID       int32
	ReplayPlayerName     [32]byte
	LevelHashMD5         [16]byte
	TimeLvl2             float32
	TimeLvl3             float32
	TimeLvl4             float32
	LeviDownTime         float32
	OrbDownTime          float32
	Status               int32
	HomingMax            int32
	TimeHomingMax        float32
	EnemiesAliveMax      int32
	TimeEnemiesAliveMax  float32
	TimeMax              float32
	Padding1             [4]byte // cause computer science
	StatsBase            int64   // address to stat frame array
	StatsFramesLoaded    int32
	StatsFinishedLoading bool
	Padding2             [3]byte // Padding here because previous data is in a struct with a pointer.
	StartingHandLevel    int32
	StartingHomingCount  int32
	StartingTime         float32
	ProhibitedMods       bool
}

type StatsFrame struct {
	GemsCollected      int32
	Kills              int32
	DaggersFired       int32
	DaggersHit         int32
	EnemiesAlive       int32
	LevelGems          int32
	HomingDaggers      int32
	GemsDespawned      int32
	GemsEaten          int32
	TotalGems          int32
	DaggersEaten       int32
	PerEnemyAliveCount [17]int16
	PerEnemyKillCount  [17]int16
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
		return fmt.Errorf("RefreshData: unable to encode data block: %w", err)
	}

	return nil
}

func (dd *DevilDaggers) refreshStatsFrame() error {
	if dd.connected != true {
		return nil
	}

	framesLoaded := int(dd.dataBlock.StatsFramesLoaded)

	dd.statsFrame = make([]StatsFrame, framesLoaded)

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
		return fmt.Errorf("RefreshStatsFrame: unable to encode data block: %w", err)
	}

	return nil

}

func (dd *DevilDaggers) GetDDStatsVersion() int32 {
	return dd.dataBlock.DDStatsVersion
}

func (dd *DevilDaggers) GetPlayerID() int32 {
	return dd.dataBlock.PlayerID
}

func (dd *DevilDaggers) GetPlayerName() string {
	return byteArrayToString(&dd.dataBlock.UserName)
}

func (dd *DevilDaggers) GetTime() float32 {
	return dd.dataBlock.Time
}

func (dd *DevilDaggers) GetGemsCollected() int32 {
	return dd.dataBlock.GemsCollected
}

func (dd *DevilDaggers) GetKills() int32 {
	return dd.dataBlock.Kills
}

func (dd *DevilDaggers) GetDaggersFired() int32 {
	return dd.dataBlock.DaggersFired
}

func (dd *DevilDaggers) GetDaggersHit() int32 {
	return dd.dataBlock.DaggersHit
}

func (dd *DevilDaggers) GetAccuracy() float32 {
	if dd.dataBlock.DaggersFired == 0 {
		return 0.0
	}
	return float32(dd.dataBlock.DaggersHit) / float32(dd.dataBlock.DaggersFired) * 100
}

func (dd *DevilDaggers) GetEnemiesAlive() int32 {
	return dd.dataBlock.EnemiesAlive
}

func (dd *DevilDaggers) GetLevelGems() int32 {
	return dd.dataBlock.LevelGems
}

func (dd *DevilDaggers) GetHomingDaggers() int32 {
	return dd.dataBlock.HomingDaggers
}

func (dd *DevilDaggers) GetGemsDespawned() int32 {
	return dd.dataBlock.GemsDespawned
}

func (dd *DevilDaggers) GetGemsEaten() int32 {
	return dd.dataBlock.GemsEaten
}

func (dd *DevilDaggers) GetTotalGems() int32 {
	return dd.dataBlock.TotalGems
}

func (dd *DevilDaggers) GetDaggersEaten() int32 {
	return dd.dataBlock.DaggersEaten
}

func (dd *DevilDaggers) GetSkull1Alive() int16 {
	return dd.dataBlock.PerEnemyAliveCount[skull1]
}

func (dd *DevilDaggers) GetSkull2Alive() int16 {
	return dd.dataBlock.PerEnemyAliveCount[skull2]
}

func (dd *DevilDaggers) GetSpiderlingAlive() int16 {
	return dd.dataBlock.PerEnemyAliveCount[spiderling]
}

func (dd *DevilDaggers) GetSkull3Alive() int16 {
	return dd.dataBlock.PerEnemyAliveCount[skull3]
}

func (dd *DevilDaggers) GetSquid1Alive() int16 {
	return dd.dataBlock.PerEnemyAliveCount[squid1]
}

func (dd *DevilDaggers) GetSquid2Alive() int16 {
	return dd.dataBlock.PerEnemyAliveCount[squid2]
}

func (dd *DevilDaggers) GetSquid3Alive() int16 {
	return dd.dataBlock.PerEnemyAliveCount[squid3]
}

func (dd *DevilDaggers) GetCentipede1Alive() int16 {
	return dd.dataBlock.PerEnemyAliveCount[centipede1]
}

func (dd *DevilDaggers) GetCentipede2Alive() int16 {
	return dd.dataBlock.PerEnemyAliveCount[centipede2]
}

func (dd *DevilDaggers) GetSpider1Alive() int16 {
	return dd.dataBlock.PerEnemyAliveCount[spider1]
}

func (dd *DevilDaggers) GetSpider2Alive() int16 {
	return dd.dataBlock.PerEnemyAliveCount[spider2]
}

func (dd *DevilDaggers) GetLeviathanAlive() int16 {
	return dd.dataBlock.PerEnemyAliveCount[leviathan]
}

func (dd *DevilDaggers) GetOrbAlive() int16 {
	return dd.dataBlock.PerEnemyAliveCount[orb]
}

func (dd *DevilDaggers) GetThornAlive() int16 {
	return dd.dataBlock.PerEnemyAliveCount[thorn]
}

func (dd *DevilDaggers) GetCentipede3Alive() int16 {
	return dd.dataBlock.PerEnemyAliveCount[centipede3]
}

func (dd *DevilDaggers) GetSpiderEggAlive() int16 {
	return dd.dataBlock.PerEnemyAliveCount[spiderEgg]
}

func (dd *DevilDaggers) GetSkull1Killed() int16 {
	return dd.dataBlock.PerEnemyKillCount[skull1]
}

func (dd *DevilDaggers) GetSkull2Killed() int16 {
	return dd.dataBlock.PerEnemyKillCount[skull2]
}

func (dd *DevilDaggers) GetSpiderlingKilled() int16 {
	return dd.dataBlock.PerEnemyKillCount[spiderling]
}

func (dd *DevilDaggers) GetSkull3Killed() int16 {
	return dd.dataBlock.PerEnemyKillCount[skull3]
}

func (dd *DevilDaggers) GetSquid1Killed() int16 {
	return dd.dataBlock.PerEnemyKillCount[squid1]
}

func (dd *DevilDaggers) GetSquid2Killed() int16 {
	return dd.dataBlock.PerEnemyKillCount[squid2]
}

func (dd *DevilDaggers) GetSquid3Killed() int16 {
	return dd.dataBlock.PerEnemyKillCount[squid3]
}

func (dd *DevilDaggers) GetCentipede1Killed() int16 {
	return dd.dataBlock.PerEnemyKillCount[centipede1]
}

func (dd *DevilDaggers) GetCentipede2Killed() int16 {
	return dd.dataBlock.PerEnemyKillCount[centipede2]
}

func (dd *DevilDaggers) GetSpider1Killed() int16 {
	return dd.dataBlock.PerEnemyKillCount[spider1]
}

func (dd *DevilDaggers) GetSpider2Killed() int16 {
	return dd.dataBlock.PerEnemyKillCount[spider2]
}

func (dd *DevilDaggers) GetLeviathanKilled() int16 {
	return dd.dataBlock.PerEnemyKillCount[leviathan]
}

func (dd *DevilDaggers) GetOrbKilled() int16 {
	return dd.dataBlock.PerEnemyKillCount[orb]
}

func (dd *DevilDaggers) GetThornKilled() int16 {
	return dd.dataBlock.PerEnemyKillCount[thorn]
}

func (dd *DevilDaggers) GetCentipede3Killed() int16 {
	return dd.dataBlock.PerEnemyKillCount[centipede3]
}

func (dd *DevilDaggers) GetSpiderEggKilled() int16 {
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

func (dd *DevilDaggers) GetReplayPlayerID() int32 {
	return dd.dataBlock.ReplayPlayerID
}

func (dd *DevilDaggers) GetReplayPlayerName() string {
	return byteArrayToString(&dd.dataBlock.ReplayPlayerName)
}

func (dd *DevilDaggers) GetLevelHashMD5() string {
	return fmt.Sprintf("%x", dd.dataBlock.LevelHashMD5)
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

func (dd *DevilDaggers) GetStatus() int32 {
	return dd.dataBlock.Status
}

func (dd *DevilDaggers) GetHomingMax() int32 {
	return dd.dataBlock.HomingMax
}

func (dd *DevilDaggers) GetHomingMaxTime() float32 {
	return dd.dataBlock.TimeHomingMax
}

func (dd *DevilDaggers) GetEnemiesAliveMax() int32 {
	return dd.dataBlock.EnemiesAliveMax
}

func (dd *DevilDaggers) GetEnemiesAliveMaxTime() float32 {
	return dd.dataBlock.TimeEnemiesAliveMax
}

func (dd *DevilDaggers) GetTimeMax() float32 {
	return dd.dataBlock.TimeMax
}

func (dd *DevilDaggers) GetStatsFramesLoaded() int32 {
	return dd.dataBlock.StatsFramesLoaded
}

func (dd *DevilDaggers) GetStatsFinishedLoading() bool {
	return dd.dataBlock.StatsFinishedLoading
}

func (dd *DevilDaggers) GetStartingHandLevel() int32 {
	return dd.dataBlock.StartingHandLevel
}

func (dd *DevilDaggers) GetStartingHomingCount() int32 {
	return dd.dataBlock.StartingHomingCount
}

func (dd *DevilDaggers) GetStartingTime() float32 {
	return dd.dataBlock.StartingTime
}

func (dd *DevilDaggers) GetProhibitedMods() bool {
	return dd.dataBlock.ProhibitedMods
}

func (dd *DevilDaggers) GetStatsFrame() ([]StatsFrame, error) {
	err := dd.refreshStatsFrame()
	if err != nil {
		return nil, fmt.Errorf("GetStatsFrame: could not refresh stats frame: %w", err)
	}
	return dd.statsFrame, nil
}

func byteArrayToString(a *[32]byte) string {
	for i, b := range a {
		if b == 0 {
			return string(a[:i])
		}
	}
	return string(a[:])
}

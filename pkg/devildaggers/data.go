package devildaggers

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"

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
	StatusTitle = iota
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
	GemsDespawned      uint32
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

// RefreshData attempts to read the Devil Daggers process memory. The data is acquired based
// on the __ddstats__ block within the game's memory. The data is then read into the dataBlock
// struct. The variables of this data can then be read using the various 'Get' methods.
func (dd *DevilDaggers) RefreshData() error {
	if dd.connected != true {
		return errors.New("RefreshData: connection to window lost")
	}

	buf, _, ok := w32.ReadProcessMemory(w32.HANDLE(dd.handle), uintptr(dd.ddstatsBlockAddress), dataSize)
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

	fmt.Printf("%+v\n", dd.dataBlock)

	return nil
}

func (dd *DevilDaggers) GetDDStatsVersion() int {
	return int(dd.dataBlock.DDStatsVersion)
}

func (dd *DevilDaggers) GetPlayerID() int {
	return int(dd.dataBlock.PlayerID)
}

func (dd *DevilDaggers) GetPlayerName() string {
	return byteArrayToString(&dd.dataBlock.UserName)
}

func (dd *DevilDaggers) GetTime() float64 {
	return float64(dd.dataBlock.Time)
}

func (dd *DevilDaggers) GetGemsCollected() int {
	return int(dd.dataBlock.GemsCollected)
}

func (dd *DevilDaggers) GetKills() int {
	return int(dd.dataBlock.Kills)
}

func (dd *DevilDaggers) GetDaggersFired() int {
	return int(dd.dataBlock.DaggersFired)
}

func (dd *DevilDaggers) GetDaggersHit() int {
	return int(dd.dataBlock.DaggersHit)
}

func (dd *DevilDaggers) GetAccuracy() float64 {
	if dd.dataBlock.DaggersFired == 0 {
		return 0.0
	}
	return float64(dd.dataBlock.DaggersHit / dd.dataBlock.DaggersFired)
}

func (dd *DevilDaggers) GetEnemiesAlive() int {
	return int(dd.dataBlock.EnemiesAlive)
}

func (dd *DevilDaggers) GetLevelGems() int {
	return int(dd.dataBlock.LevelGems)
}

func (dd *DevilDaggers) GetHomingDaggers() int {
	return int(dd.dataBlock.HomingDaggers)
}

func (dd *DevilDaggers) GetGemsDespawned() int {
	return int(dd.dataBlock.GemsDespawned)
}

func (dd *DevilDaggers) GetGemsEaten() int {
	return int(dd.dataBlock.GemsEaten)
}

func (dd *DevilDaggers) GetTotalGems() int {
	return int(dd.dataBlock.TotalGems)
}

func (dd *DevilDaggers) GetSkull1Alive() int {
	return int(dd.dataBlock.PerEnemyAliveCount[skull1])
}

func (dd *DevilDaggers) GetSkull2Alive() int {
	return int(dd.dataBlock.PerEnemyAliveCount[skull2])
}

func (dd *DevilDaggers) GetSpiderlingAlive() int {
	return int(dd.dataBlock.PerEnemyAliveCount[spiderling])
}

func (dd *DevilDaggers) GetSkull3Alive() int {
	return int(dd.dataBlock.PerEnemyAliveCount[skull3])
}

func (dd *DevilDaggers) GetSquid1Alive() int {
	return int(dd.dataBlock.PerEnemyAliveCount[squid1])
}

func (dd *DevilDaggers) GetSquid2Alive() int {
	return int(dd.dataBlock.PerEnemyAliveCount[squid2])
}

func (dd *DevilDaggers) GetSquid3Alive() int {
	return int(dd.dataBlock.PerEnemyAliveCount[squid3])
}

func (dd *DevilDaggers) GetCentipede1Alive() int {
	return int(dd.dataBlock.PerEnemyAliveCount[centipede1])
}

func (dd *DevilDaggers) GetCentipede2Alive() int {
	return int(dd.dataBlock.PerEnemyAliveCount[centipede2])
}

func (dd *DevilDaggers) GetSpider1Alive() int {
	return int(dd.dataBlock.PerEnemyAliveCount[spider1])
}

func (dd *DevilDaggers) GetSpider2Alive() int {
	return int(dd.dataBlock.PerEnemyAliveCount[spider2])
}

func (dd *DevilDaggers) GetLeviathanAlive() int {
	return int(dd.dataBlock.PerEnemyAliveCount[leviathan])
}

func (dd *DevilDaggers) GetOrbAlive() int {
	return int(dd.dataBlock.PerEnemyAliveCount[orb])
}

func (dd *DevilDaggers) GetThornAlive() int {
	return int(dd.dataBlock.PerEnemyAliveCount[thorn])
}

func (dd *DevilDaggers) GetCentipede3Alive() int {
	return int(dd.dataBlock.PerEnemyAliveCount[centipede3])
}

func (dd *DevilDaggers) GetSpiderEggAlive() int {
	return int(dd.dataBlock.PerEnemyAliveCount[spiderEgg])
}

func (dd *DevilDaggers) GetSkull1Killed() int {
	return int(dd.dataBlock.PerEnemyKillCount[skull1])
}

func (dd *DevilDaggers) GetSkull2Killed() int {
	return int(dd.dataBlock.PerEnemyKillCount[skull2])
}

func (dd *DevilDaggers) GetSpiderlingKilled() int {
	return int(dd.dataBlock.PerEnemyKillCount[spiderling])
}

func (dd *DevilDaggers) GetSkull3Killed() int {
	return int(dd.dataBlock.PerEnemyKillCount[skull3])
}

func (dd *DevilDaggers) GetSquid1Killed() int {
	return int(dd.dataBlock.PerEnemyKillCount[squid1])
}

func (dd *DevilDaggers) GetSquid2Killed() int {
	return int(dd.dataBlock.PerEnemyKillCount[squid2])
}

func (dd *DevilDaggers) GetSquid3Killed() int {
	return int(dd.dataBlock.PerEnemyKillCount[squid3])
}

func (dd *DevilDaggers) GetCentipede1Killed() int {
	return int(dd.dataBlock.PerEnemyKillCount[centipede1])
}

func (dd *DevilDaggers) GetCentipede2Killed() int {
	return int(dd.dataBlock.PerEnemyKillCount[centipede2])
}

func (dd *DevilDaggers) GetSpider1Killed() int {
	return int(dd.dataBlock.PerEnemyKillCount[spider1])
}

func (dd *DevilDaggers) GetSpider2Killed() int {
	return int(dd.dataBlock.PerEnemyKillCount[spider2])
}

func (dd *DevilDaggers) GetLeviathanKilled() int {
	return int(dd.dataBlock.PerEnemyKillCount[leviathan])
}

func (dd *DevilDaggers) GetOrbKilled() int {
	return int(dd.dataBlock.PerEnemyKillCount[orb])
}

func (dd *DevilDaggers) GetThornKilled() int {
	return int(dd.dataBlock.PerEnemyKillCount[thorn])
}

func (dd *DevilDaggers) GetCentipede3Killed() int {
	return int(dd.dataBlock.PerEnemyKillCount[centipede3])
}

func (dd *DevilDaggers) GetSpiderEggKilled() int {
	return int(dd.dataBlock.PerEnemyKillCount[spiderEgg])
}

func (dd *DevilDaggers) GetIsPlayerAlive() bool {
	return dd.dataBlock.IsPlayerAlive
}

func (dd *DevilDaggers) GetIsReplay() bool {
	return dd.dataBlock.IsReplay
}

func (dd *DevilDaggers) GetDeathType() int {
	return int(dd.dataBlock.DeathType)
}

func (dd *DevilDaggers) GetIsInGame() bool {
	return dd.dataBlock.IsInGame
}

func (dd *DevilDaggers) GetReplayPlayerID() int {
	return int(dd.dataBlock.ReplayPlayerID)
}

func (dd *DevilDaggers) GetReplayPlayerName() string {
	return byteArrayToString(&dd.dataBlock.ReplayPlayerName)
}

func (dd *DevilDaggers) GetLevelHashMD5() string {
	return fmt.Sprintf("%x", dd.dataBlock.LevelHashMD5)
}

func (dd *DevilDaggers) GetTimeLvl2() float64 {
	return float64(dd.dataBlock.TimeLvl2)
}

func (dd *DevilDaggers) GetTimeLvl3() float64 {
	return float64(dd.dataBlock.TimeLvl3)
}

func (dd *DevilDaggers) GetTimeLvl4() float64 {
	return float64(dd.dataBlock.TimeLvl4)
}

func (dd *DevilDaggers) GetLeviathanDownTime() float64 {
	return float64(dd.dataBlock.LeviDownTime)
}

func (dd *DevilDaggers) GetOrbDownTime() float64 {
	return float64(dd.dataBlock.OrbDownTime)
}

func (dd *DevilDaggers) GetStatus() int {
	return int(dd.dataBlock.Status)
}

func byteArrayToString(a *[32]byte) string {
	for i, b := range a {
		if b == 0 {
			return string(a[:i])
		}
	}
	return string(a[:])
}

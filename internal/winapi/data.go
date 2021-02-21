package ddstats

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

func (dd *DDStats) RefreshDevilDaggersData() error {
	if dd.connected != true {
		return errors.New("RefreshDevilDaggersDataBlock: connection to window lost")
	}

	buf, _, ok := w32.ReadProcessMemory(w32.HANDLE(dd.handle), uintptr(dd.ddstatsBlockAddress), dataSize)
	if !ok {
		return errors.New("RefreshDevilDaggersDataBlock: unable to read process memory")
	}

	byteBuf := bytes.NewBuffer(make([]byte, 0, len(buf)*2))

	for _, b := range buf {
		split := make([]byte, 2)
		binary.LittleEndian.PutUint16(split, b)
		byteBuf.Write(split)
	}

	err := binary.Read(byteBuf, binary.LittleEndian, dd.dataBlock)
	if err != nil {
		return fmt.Errorf("RefreshDevilDaggersDataBlock: unable to encode data block: %w", err)
	}

	fmt.Printf("%+v", dd.dataBlock)

	return nil
}

func (dd *DDStats) GetDDStatsVersion() int {
	return int(dd.dataBlock.DDStatsVersion)
}

func (dd *DDStats) GetPlayerID() int {
	return int(dd.dataBlock.PlayerID)
}

func (dd *DDStats) GetPlayerName() string {
	return string(dd.dataBlock.UserName[:])
}

func (dd *DDStats) GetTime() float64 {
	return float64(dd.dataBlock.Time)
}

func (dd *DDStats) GetGemsCollected() int {
	return int(dd.dataBlock.GemsCollected)
}

func (dd *DDStats) GetKills() int {
	return int(dd.dataBlock.Kills)
}

func (dd *DDStats) GetDaggersFired() int {
	return int(dd.dataBlock.DaggersFired)
}

func (dd *DDStats) GetDaggersHit() int {
	return int(dd.dataBlock.DaggersHit)
}

func (dd *DDStats) GetAccuracy() float64 {
	if dd.dataBlock.DaggersFired == 0 {
		return 0.0
	}
	return float64(dd.dataBlock.DaggersHit / dd.dataBlock.DaggersFired)
}

func (dd *DDStats) GetEnemiesAlive() int {
	return int(dd.dataBlock.EnemiesAlive)
}

func (dd *DDStats) GetLevelGems() int {
	return int(dd.dataBlock.LevelGems)
}

func (dd *DDStats) GetHomingDaggers() int {
	return int(dd.dataBlock.HomingDaggers)
}

func (dd *DDStats) GetGemsDespawned() int {
	return int(dd.dataBlock.GemsDespawned)
}

func (dd *DDStats) GetGemsEaten() int {
	return int(dd.dataBlock.GemsEaten)
}

func (dd *DDStats) GetTotalGems() int {
	return int(dd.dataBlock.TotalGems)
}

func (dd *DDStats) GetSkull1Alive() int {
	return int(dd.dataBlock.PerEnemyAliveCount[skull1])
}

func (dd *DDStats) GetSkull2Alive() int {
	return int(dd.dataBlock.PerEnemyAliveCount[skull2])
}

func (dd *DDStats) GetSpiderlingAlive() int {
	return int(dd.dataBlock.PerEnemyAliveCount[spiderling])
}

func (dd *DDStats) GetSkull3Alive() int {
	return int(dd.dataBlock.PerEnemyAliveCount[skull3])
}

func (dd *DDStats) GetSquid1Alive() int {
	return int(dd.dataBlock.PerEnemyAliveCount[squid1])
}

func (dd *DDStats) GetSquid2Alive() int {
	return int(dd.dataBlock.PerEnemyAliveCount[squid2])
}

func (dd *DDStats) GetSquid3Alive() int {
	return int(dd.dataBlock.PerEnemyAliveCount[squid3])
}

func (dd *DDStats) GetCentipede1Alive() int {
	return int(dd.dataBlock.PerEnemyAliveCount[centipede1])
}

func (dd *DDStats) GetCentipede2Alive() int {
	return int(dd.dataBlock.PerEnemyAliveCount[centipede2])
}

func (dd *DDStats) GetSpider1Alive() int {
	return int(dd.dataBlock.PerEnemyAliveCount[spider1])
}

func (dd *DDStats) GetSpider2Alive() int {
	return int(dd.dataBlock.PerEnemyAliveCount[spider2])
}

func (dd *DDStats) GetLeviathanAlive() int {
	return int(dd.dataBlock.PerEnemyAliveCount[leviathan])
}

func (dd *DDStats) GetOrbAlive() int {
	return int(dd.dataBlock.PerEnemyAliveCount[orb])
}

func (dd *DDStats) GetThornAlive() int {
	return int(dd.dataBlock.PerEnemyAliveCount[thorn])
}

func (dd *DDStats) GetCentipede3Alive() int {
	return int(dd.dataBlock.PerEnemyAliveCount[centipede3])
}

func (dd *DDStats) GetSpiderEggAlive() int {
	return int(dd.dataBlock.PerEnemyAliveCount[spiderEgg])
}

func (dd *DDStats) GetSkull1Killed() int {
	return int(dd.dataBlock.PerEnemyKillCount[skull1])
}

func (dd *DDStats) GetSkull2Killed() int {
	return int(dd.dataBlock.PerEnemyKillCount[skull2])
}

func (dd *DDStats) GetSpiderlingKilled() int {
	return int(dd.dataBlock.PerEnemyKillCount[spiderling])
}

func (dd *DDStats) GetSkull3Killed() int {
	return int(dd.dataBlock.PerEnemyKillCount[skull3])
}

func (dd *DDStats) GetSquid1Killed() int {
	return int(dd.dataBlock.PerEnemyKillCount[squid1])
}

func (dd *DDStats) GetSquid2Killed() int {
	return int(dd.dataBlock.PerEnemyKillCount[squid2])
}

func (dd *DDStats) GetSquid3Killed() int {
	return int(dd.dataBlock.PerEnemyKillCount[squid3])
}

func (dd *DDStats) GetCentipede1Killed() int {
	return int(dd.dataBlock.PerEnemyKillCount[centipede1])
}

func (dd *DDStats) GetCentipede2Killed() int {
	return int(dd.dataBlock.PerEnemyKillCount[centipede2])
}

func (dd *DDStats) GetSpider1Killed() int {
	return int(dd.dataBlock.PerEnemyKillCount[spider1])
}

func (dd *DDStats) GetSpider2Killed() int {
	return int(dd.dataBlock.PerEnemyKillCount[spider2])
}

func (dd *DDStats) GetLeviathanKilled() int {
	return int(dd.dataBlock.PerEnemyKillCount[leviathan])
}

func (dd *DDStats) GetOrbKilled() int {
	return int(dd.dataBlock.PerEnemyKillCount[orb])
}

func (dd *DDStats) GetThornKilled() int {
	return int(dd.dataBlock.PerEnemyKillCount[thorn])
}

func (dd *DDStats) GetCentipede3Killed() int {
	return int(dd.dataBlock.PerEnemyKillCount[centipede3])
}

func (dd *DDStats) GetSpiderEggKilled() int {
	return int(dd.dataBlock.PerEnemyKillCount[spiderEgg])
}

func (dd *DDStats) GetIsPlayerAlive() bool {
	return dd.dataBlock.IsPlayerAlive
}

func (dd *DDStats) GetIsReplay() bool {
	return dd.dataBlock.IsReplay
}

func (dd *DDStats) GetDeathType() int {
	return int(dd.dataBlock.DeathType)
}

func (dd *DDStats) GetIsInGame() bool {
	return dd.dataBlock.IsInGame
}

func (dd *DDStats) GetReplayPlayerID() int {
	return int(dd.dataBlock.ReplayPlayerID)
}

func (dd *DDStats) GetReplayPlayerName() string {
	return string(dd.dataBlock.ReplayPlayerName[:])
}

func (dd *DDStats) GetLevelHashMD5() string {
	return fmt.Sprintf("%x", dd.dataBlock.LevelHashMD5)
}

func (dd *DDStats) GetTimeLvl2() float64 {
	return float64(dd.dataBlock.TimeLvl2)
}

func (dd *DDStats) GetTimeLvl3() float64 {
	return float64(dd.dataBlock.TimeLvl3)
}

func (dd *DDStats) GetTimeLvl4() float64 {
	return float64(dd.dataBlock.TimeLvl4)
}

func (dd *DDStats) GetLeviathanDownTime() float64 {
	return float64(dd.dataBlock.LeviDownTime)
}

func (dd *DDStats) GetOrbDownTime() float64 {
	return float64(dd.dataBlock.OrbDownTime)
}

func (dd *DDStats) GetStatus() int {
	return int(dd.dataBlock.Status)
}

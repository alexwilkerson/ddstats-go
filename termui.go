package main

import (
	"fmt"
	"time"

	"github.com/TheTitanrain/w32"
	ui "github.com/gizak/termui"
)

type statDisplay struct {
	timer         float32
	daggersHit    int
	daggersFired  int
	accuracy      float64
	totalGems     int
	homing        int
	enemiesAlive  int
	enemiesKilled int
}

const logoString = `@@@@@@@   @@@@@@@    @@@@@@   @@@@@@@   @@@@@@   @@@@@@@   @@@@@@
@@@@@@@@  @@@@@@@@  @@@@@@@   @@@@@@@  @@@@@@@@  @@@@@@@  @@@@@@@
@@!  @@@  @@!  @@@  !@@         @@!    @@!  @@@    @@!    !@@
!@!  @!@  !@!  @!@  !@!         !@!    !@!  @!@    !@!    !@!
@!@  !@!  @!@  !@!  !!@@!!      @!!    @!@!@!@!    @!!    !!@@!!
!@!  !!!  !@!  !!!   !!@!!!     !!!    !!!@!!!!    !!!     !!@!!!
!!:  !!!  !!:  !!!       !:!    !!:    !!:  !!!    !!:         !:!
:!:  !:!  :!:  !:!      !:!     :!:    :!:  !:!    :!:        !:!
 :::: ::   :::: ::  :::: ::      ::    ::   :::     ::    :::: ::
:: :  :   :: :  :   :: : :       :      :   : :     :     :: : :`

func (sd *statDisplay) Update() {
	if gc.status == statusInMainMenu || gc.status == statusInDaggerLobby || gc.status == statusConnecting || gc.status == statusNotConnected {
		sd.Reset()
	} else {
		sd.timer = gc.timer
		sd.daggersHit = gc.daggersHit
		sd.daggersFired = gc.daggersFired
		sd.accuracy = gc.accuracy
		if gc.GetStatus() == statusIsPlaying {
			sd.totalGems = -1
			sd.homing = -1
		} else if gc.GetStatus() == statusIsDead {
			sd.totalGems = gc.totalGemsAtDeath
			sd.homing = gc.homingAtDeath
		} else {
			sd.totalGems = gc.totalGems
			sd.homing = gc.homing
		}
		sd.enemiesAlive = gc.enemiesAlive
		sd.enemiesKilled = gc.enemiesKilled
	}
}

func (sd *statDisplay) Reset() {
	sd.timer = 0.0
	sd.daggersHit = 0
	sd.daggersFired = 0
	sd.accuracy = 0.0
	sd.totalGems = 0
	sd.homing = 0
	sd.enemiesAlive = 0
	sd.enemiesKilled = 0
}

func setupHandles() {
	ui.Handle("<f10>", quit)
}

func quit(ui.Event) {
	w32.CloseHandle(handle)
	ui.StopLoop()
}

func classicLayout() {
	logo := ui.NewPar(logoString)
	logo.TextFgColor = ui.StringToAttribute("red")
	logo.Border = false
	logo.SetX(ui.TermWidth()/2 - 34)
	logo.SetY(1)
	logo.Width = 67
	logo.Height = 10

	nameLabel := ui.NewPar("")
	nameLabel.TextFgColor = ui.StringToAttribute("red")
	nameLabel.Border = false
	nameLabel.X = ui.TermWidth()/2 - 34
	nameLabel.Y = 11
	nameLabel.Width = 34
	nameLabel.Height = 1

	versionLabel := ui.NewPar("v" + version)
	versionLabel.TextFgColor = ui.StringToAttribute("red")
	versionLabel.Border = false
	versionLabel.X = ui.TermWidth()/2 + 31 - (len(version) + 1)
	versionLabel.Y = 11
	versionLabel.Width = len(version) + 1
	versionLabel.Height = 1

	exitLabel := ui.NewPar("[F10] Exit")
	exitLabel.Border = false
	exitLabel.X = ui.TermWidth()/2 - 34
	exitLabel.Y = 21
	exitLabel.Height = 1
	exitLabel.Width = 10

	updateLabel := ui.NewPar("(UPDATE AVAILABLE)")
	updateLabel.TextFgColor = ui.StringToAttribute("green")
	updateLabel.Border = false
	updateLabel.X = ui.TermWidth()/2 - 34
	updateLabel.Y = 11
	updateLabel.Width = 18
	updateLabel.Height = 1

	motdLabel := ui.NewPar("Fetching MOTD.")
	motdLabel.X = ui.TermWidth()/2 - 7
	motdLabel.Border = false
	motdLabel.Y = 12
	motdLabel.Height = 1
	motdLabel.Width = 14

	statusLabel := ui.NewPar("")
	statusLabel.Border = false
	statusLabel.X = ui.TermWidth() / 2
	statusLabel.Y = 13
	statusLabel.Height = 1
	statusLabel.Width = 70

	statsLeft := ui.NewPar("")
	statsLeft.SetX(ui.TermWidth()/2 - 34)
	statsLeft.SetY(15)
	statsLeft.Border = false
	statsLeft.Width = 34
	statsLeft.Height = 5

	statsRight := ui.NewPar("")
	statsRight.SetX(ui.TermWidth() / 2)
	statsRight.SetY(15)
	statsRight.Border = false
	statsRight.Width = 34
	statsRight.Height = 5

	ui.Render(logo, versionLabel, exitLabel)

	for {

		gc.GetGameVariables()
		sd.Update()

		nameLabel.Text = fmt.Sprintf("%v", gc.playerName)

		if updateAvailable {
			ui.Render(updateLabel)
		}

		if motd != "" {
			motdLabel.X = ui.TermWidth()/2 - len(motd)/2
			motdLabel.Width = len(motd) + 1
			motdLabel.Text = motd
		}

		var statusString string
		switch gc.GetStatus() {
		case statusNotConnected:
			statusString = "Devil Daggers not found"
			statusLabel.TextFgColor = ui.StringToAttribute("red")
		case statusInMainMenu:
			statusString = "In main menu"
			statusLabel.TextFgColor = ui.StringToAttribute("green")
		case statusInDaggerLobby:
			statusString = "In dagger lobby"
			statusLabel.TextFgColor = ui.StringToAttribute("green")
		case statusIsReplay:
			statusString = "Watching replay"
			statusLabel.TextFgColor = ui.StringToAttribute("green")
		case statusIsPlaying:
			statusString = "Currently playing"
			statusLabel.TextFgColor = ui.StringToAttribute("green")
		case statusConnecting:
			statusString = "Connecting to Devil Daggers"
			statusLabel.TextFgColor = ui.StringToAttribute("yellow")
		case statusIsDead:
			statusString = "Death screen"
			statusLabel.TextFgColor = ui.StringToAttribute("red")
		}
		statusLabel.X = ui.TermWidth()/2 - (len(statusString)/2 + 6)

		statusLabel.X = ui.TermWidth()/2 - len(statusString)/2 - 19
		statusLabel.Height = 1
		statusLabel.Text = "                [[ " + statusString + " ]]                "

		ui.Render(nameLabel, motdLabel, statusLabel)

		timerString := fmt.Sprintf("In Game Timer: %.4fs", sd.timer)
		daggersHitString := fmt.Sprintf("Daggers Hit: %d", sd.daggersHit)
		daggersFiredString := fmt.Sprintf("Daggers Fired: %d", sd.daggersFired)
		accuracyString := fmt.Sprintf("Accuracy: %.2f%%", sd.accuracy)
		var gemsString string
		if sd.totalGems == -1 {
			gemsString = "Gems: HIDDEN"
		} else {
			gemsString = fmt.Sprintf("Gems: %d", sd.totalGems)
		}
		var homingString string
		if sd.homing == -1 {
			homingString = "Homing Daggers: HIDDEN"
		} else {
			homingString = fmt.Sprintf("Homing Daggers: %d", sd.homing)
		}
		enemiesAliveString := fmt.Sprintf("Enemies Alive: %d", sd.enemiesAlive)
		enemiesKilledString := fmt.Sprintf("Enemies Killed: %d", sd.enemiesKilled)

		statsLeft.Text = fmt.Sprintf("%v\n%v\n%v\n%v\n", timerString, daggersHitString, daggersFiredString, accuracyString)
		statsRight.Text = fmt.Sprintf("%32v\n%32v\n%32v\n%32v\n", gemsString, homingString, enemiesAliveString, enemiesKilledString)

		ui.Render(statsLeft, statsRight)
		time.Sleep(time.Second / captureFPS)
	}

}

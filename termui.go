package main

import (
	"fmt"
	"time"

	"github.com/TheTitanrain/w32"
	ui "github.com/gizak/termui"
)

type statDisplay struct {
	deathScreen   bool
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
	if gameCapture.status == statusInMainMenu || gameCapture.status == statusInDaggerLobby || gameCapture.status == statusConnecting || gameCapture.status == statusNotConnected {
		sd.Reset()
	} else {
		sd.timer = gameCapture.timer
		sd.daggersHit = gameCapture.daggersHit
		sd.daggersFired = gameCapture.daggersFired
		sd.accuracy = gameCapture.accuracy
		if gameCapture.GetStatus() == statusIsPlaying {
			sd.totalGems = -1
			sd.homing = -1
		} else if gameCapture.GetStatus() == statusIsDead {
			sd.totalGems = gameCapture.totalGemsAtDeath
			sd.homing = gameCapture.homingAtDeath
		} else {
			sd.totalGems = gameCapture.totalGems
			sd.homing = gameCapture.homing
		}
		sd.enemiesAlive = gameCapture.enemiesAlive
		sd.enemiesKilled = gameCapture.enemiesKilled
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
	ui.Handle("<f9>", toggleDebug)
}

func toggleDebug(ui.Event) {
	ui.Clear()
	debugWindowVisible = !debugWindowVisible
}

func quit(ui.Event) {
	w32.CloseHandle(handle)
	ui.StopLoop()
}

func classicLayout() {
	debugWindow := ui.NewPar(debug.log)
	debugWindow.TextFgColor = ui.StringToAttribute("black")
	debugWindow.TextBgColor = ui.StringToAttribute("white")
	debugWindow.Border = true
	debugWindow.SetX(ui.TermWidth()/2 - 34)
	debugWindow.SetY(1)
	debugWindow.Width = 67
	debugWindow.Height = ui.TermHeight()
	debugWindow.Float = ui.AlignBottom

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
	updateLabel.X = ui.TermWidth()/2 - 9
	updateLabel.Y = 11
	updateLabel.Width = 19
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

	lastGameLabel := ui.NewPar("Last Submission: " + lastGameURL)
	lastGameLabel.Border = false
	lastGameLabel.X = ui.TermWidth()/2 - 34
	lastGameLabel.Y = 20
	lastGameLabel.Height = 1
	lastGameLabel.Width = 66

	for {
		if gameCapture.GetStatus() != statusIsDead {
			sd.Update()
		}

		if debugWindowVisible {
			debugWindow.Text = debug.log
			ui.Render(debugWindow)
		} else {
			ui.Render(logo, versionLabel, exitLabel)

			nameLabel.Text = fmt.Sprintf("%v", gameCapture.playerName)

			if motd != "" {
				motdLabel.X = ui.TermWidth()/2 - len(motd)/2
				motdLabel.Width = len(motd) + 1
				motdLabel.Text = motd
			}

			var statusString string
			switch gameCapture.GetStatus() {
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

			if updateAvailable {
				ui.Render(updateLabel)
			}

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

			lastGameLabel.Text = "Last Submission: " + lastGameURL

			ui.Render(statsLeft, statsRight, lastGameLabel)
		}
		time.Sleep(time.Second / uiFPS)
	}

}

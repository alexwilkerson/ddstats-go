package main

import (
	"fmt"
	"time"

	"github.com/TheTitanrain/w32"
	"github.com/atotto/clipboard"
	ui "github.com/gizak/termui"
)

var lastGameURLCopyTime time.Time

type StatDisplay struct {
	timer         float32
	daggersHit    int32
	daggersFired  int32
	accuracy      float32
	totalGems     int32
	homing        int32
	enemiesAlive  int32
	enemiesKilled int32
	deathType     int32
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

func (sd *StatDisplay) Update() {
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
		sd.deathType = gameCapture.deathType
	}
}

func (sd *StatDisplay) Reset() {
	sd.timer = 0.0
	sd.daggersHit = 0
	sd.daggersFired = 0
	sd.accuracy = 0.0
	sd.totalGems = 0
	sd.homing = 0
	sd.enemiesAlive = 0
	sd.enemiesKilled = 0
	sd.deathType = 0
}

func setupHandles() {
	ui.Handle("<f10>", quit)
	ui.Handle("<f9>", toggleDebug)
	ui.Handle("<MouseLeft>", copyGameURLToClipboard)
}

func toggleDebug(ui.Event) {
	ui.Clear()
	debugWindowVisible = !debugWindowVisible
}

func quit(ui.Event) {
	w32.CloseHandle(handle)
	ui.StopLoop()
}

func copyGameURLToClipboard(ui.Event) {
	if lastGameURL[:5] == "https" {
		lastGameURLCopyTime = time.Now()
		clipboard.WriteAll(lastGameURL)
	}
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
	updateLabel.X = ui.TermWidth()/2 - 34
	updateLabel.Y = 0
	updateLabel.Width = 19
	updateLabel.Height = 1

	motdLabel := ui.NewPar("")
	motdLabel.X = ui.TermWidth()/2 - 7
	motdLabel.Border = false
	motdLabel.Y = 12
	motdLabel.Height = 1
	motdLabel.Width = 14
	if config.offlineMode || !config.getMOTD {
		motdLabel.Text = ""
	} else {
		motdLabel.Text = "Fetching MOTD."
	}

	statusLabel := ui.NewPar("")
	statusLabel.Border = false
	statusLabel.X = ui.TermWidth() / 2
	statusLabel.Y = 13
	statusLabel.Height = 1
	statusLabel.Width = 70

	onlineLabel := ui.NewPar("")
	onlineLabel.Border = false
	onlineLabel.X = ui.TermWidth() / 2
	onlineLabel.Y = 11
	onlineLabel.Height = 1
	onlineLabel.Width = 20

	recordingLabel := ui.NewPar("")
	recordingLabel.Border = false
	recordingLabel.X = ui.TermWidth() / 2
	recordingLabel.Y = 14
	recordingLabel.Height = 1
	recordingLabel.Width = 20

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
			statDisplay.Update()
		}

		if debugWindowVisible {
			debugWindow.Text = debug.log
			ui.Render(debugWindow)
		} else {
			ui.Render(logo, versionLabel, exitLabel)

			if gameCapture.GetStatus() == statusNotConnected || gameCapture.GetStatus() == statusConnecting {
				nameLabel.Text = "                           "
			} else {
				nameLabel.Text = fmt.Sprintf("%v", gameCapture.playerName)
			}
			nameLabel.Width = len(nameLabel.Text) + 1

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
				statusString = deathTypes[gameCapture.deathType]
				statusLabel.TextFgColor = ui.StringToAttribute("red")
			}

			statusLabel.X = ui.TermWidth()/2 - len(statusString)/2 - 19
			statusLabel.Text = "                [[ " + statusString + " ]]                "

			if gameCapture.GetStatus() == statusNotConnected || gameCapture.GetStatus() == statusConnecting {
				onlineLabel.Text = "                    "
			} else {
				switch sioVariables.status {
				case sioStatusDisconnected:
					onlineLabel.TextFgColor = ui.StringToAttribute("red")
					onlineLabel.Text = "[[ Disconnected ]]"
				case sioStatusConnecting:
					onlineLabel.TextFgColor = ui.StringToAttribute("yellow")
					onlineLabel.Text = "[[ Connecting... ]]"
				case sioStatusTimeout:
					onlineLabel.TextFgColor = ui.StringToAttribute("red")
					onlineLabel.Text = "  [[ Timed out ]]  "
				case sioStatusLoggedIn:
					onlineLabel.TextFgColor = ui.StringToAttribute("green")
					onlineLabel.Text = "  [[ Logged in ]]  "
				case sioStatusConnected:
					onlineLabel.TextFgColor = ui.StringToAttribute("green")
					onlineLabel.Text = "  [[ Connected ]]  "
				}
			}
			onlineLabel.X = ui.TermWidth()/2 - len(onlineLabel.Text)/2

			if gameCapture.GetStatus() == statusIsPlaying || gameCapture.GetStatus() == statusIsReplay {
				recordingLabel.TextFgColor = ui.StringToAttribute("bold, green")
				recordingLabel.Text = "  [[ Recording ]]  "
			} else {
				recordingLabel.TextFgColor = ui.StringToAttribute("red")
				recordingLabel.Text = "[[ Not recording ]]"
			}
			recordingLabel.X = ui.TermWidth()/2 - len(recordingLabel.Text)/2

			ui.Render(nameLabel, motdLabel, onlineLabel, statusLabel, recordingLabel)

			if updateAvailable {
				ui.Render(updateLabel)
			}

			timerString := fmt.Sprintf("In Game Timer: %.4fs", statDisplay.timer)
			daggersHitString := fmt.Sprintf("Daggers Hit: %d", statDisplay.daggersHit)
			daggersFiredString := fmt.Sprintf("Daggers Fired: %d", statDisplay.daggersFired)
			accuracyString := fmt.Sprintf("Accuracy: %.2f%%", statDisplay.accuracy)
			var gemsString string
			if statDisplay.totalGems == -1 {
				gemsString = "Gems: HIDDEN"
			} else {
				gemsString = fmt.Sprintf("Gems: %d", statDisplay.totalGems)
			}
			var homingString string
			if statDisplay.homing == -1 {
				homingString = "Homing Daggers: HIDDEN"
			} else {
				homingString = fmt.Sprintf("Homing Daggers: %d", statDisplay.homing)
			}
			enemiesAliveString := fmt.Sprintf("Enemies Alive: %d", statDisplay.enemiesAlive)
			enemiesKilledString := fmt.Sprintf("Enemies Killed: %d", statDisplay.enemiesKilled)

			statsLeft.Text = fmt.Sprintf("%v\n%v\n%v\n%v\n", timerString, daggersHitString, daggersFiredString, accuracyString)
			statsRight.Text = fmt.Sprintf("%32v\n%32v\n%32v\n%32v\n", gemsString, homingString, enemiesAliveString, enemiesKilledString)

			if time.Since(lastGameURLCopyTime).Seconds() < 1.5 {
				lastGameLabel.Text = "Last Submission: (copied to clipboard)"
			} else {
				lastGameLabel.Text = "Last Submission: " + lastGameURL
			}

			ui.Render(statsLeft, statsRight, lastGameLabel)
		}
		time.Sleep(time.Second / uiFPS)
	}

}

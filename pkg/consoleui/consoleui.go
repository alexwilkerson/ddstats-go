package consoleui

import (
	"fmt"

	"github.com/alexwilkerson/ddstats-go/pkg/config"
	"github.com/alexwilkerson/ddstats-go/pkg/devildaggers"
	ui "github.com/gizak/termui"
)

const (
	StatusTitleScreen = iota
	StatusMenu
	StatusLobby
	StatusPlaying
	StatusDead
	StatusOwnReplayFromLastRun
	StatusOwnReplayFromLeaderboard
	StatusOtherReplay
	StatusConnecting
	StatusDevilDaggersNotFound
)

const (
	OnlineStatusDisconnected = iota
	OnlineStatusConnecting
	OnlineStatusTimedOut
	OnlineStatusLoggedIn
	OnlineStatusConnected
)

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

type Data struct {
	Host            string
	PlayerName      string
	Version         string
	UpdateAvailable bool
	MOTD            string
	Status          int32
	OnlineStatus    int
	Recording       bool
	Timer           float32
	DaggersHit      int32
	DaggersFired    int32
	Accuracy        float32
	TotalGems       int32
	Homing          int32
	EnemiesAlive    int32
	EnemiesKilled   int32
	DeathType       uint8
	LastGameID      int
}

type ConsoleUI struct {
	data *Data
}

func New(data *Data) (*ConsoleUI, error) {
	err := ui.Init()
	if err != nil {
		return nil, fmt.Errorf("New: unable to initialize termui: %w", err)
	}

	return &ConsoleUI{data: data}, nil
}

func (cui *ConsoleUI) Close() {
	ui.Close()
}

func (cui *ConsoleUI) ClearScreen() {
	ui.Clear()
}

func (cui *ConsoleUI) DrawScreen() error {
	cui.drawLogo()
	cui.drawName()
	cui.drawVersion()
	cui.drawMenu()
	if cui.data.UpdateAvailable {
		cui.drawUpdateAvailable()
	}
	cui.drawMOTD()
	err := cui.drawStatus()
	if err != nil {
		return fmt.Errorf("DrawScreen: error drawing status: %w", err)
	}
	cui.drawOnlineStatus()
	cui.drawRecording()
	cui.drawLeftSideStats()
	cui.drawRightSideStats()
	cui.drawLastGameLabel()

	return nil
}

func (cui *ConsoleUI) PollEvents() {
	uiEvents := ui.PollEvents()
	for {
		e := <-uiEvents
		switch e.ID {
		case "q", "<C-c>", "<f9>":
			return
		case "<f11>":
			config.WriteDefaultConfigFile()
		case "<MouseLeft>":
			copyGameURLToClipboard()
		}
	}
}

func copyGameURLToClipboard() {
	// if lastGameURL[:4] == "https" {
	// 	lastGameURLCopyTime = time.Now()
	// 	clipboard.WriteAll(lastGameURL)
	// }
}

func (cui *ConsoleUI) drawLogo() {
	logo := ui.NewParagraph(logoString)
	logo.TextFgColor = ui.StringToAttribute("red")
	logo.Border = false
	logo.SetX(ui.TermWidth()/2 - 34)
	logo.SetY(1)
	logo.Width = 67
	logo.Height = 10

	ui.Render(logo)
}

func (cui *ConsoleUI) drawName() {
	nameLabel := ui.NewParagraph(cui.data.PlayerName)
	nameLabel.TextFgColor = ui.StringToAttribute("red")
	nameLabel.Border = false
	nameLabel.X = ui.TermWidth()/2 - 34
	nameLabel.Y = 11
	nameLabel.Width = len(cui.data.PlayerName)
	nameLabel.Height = 1

	ui.Render(nameLabel)
}

func (cui *ConsoleUI) drawVersion() {
	versionLabel := ui.NewParagraph("v" + cui.data.Version)
	versionLabel.TextFgColor = ui.StringToAttribute("red")
	versionLabel.Border = false
	versionLabel.X = ui.TermWidth()/2 + 31 - (len(cui.data.Version) + 1)
	versionLabel.Y = 11
	versionLabel.Width = len(cui.data.Version) + 1
	versionLabel.Height = 1

	ui.Render(versionLabel)
}

func (cui *ConsoleUI) drawMenu() {
	menu := ui.NewParagraph("[F10] Exit | [F12] Reset Config File")
	menu.Border = false
	menu.X = ui.TermWidth()/2 - 34
	menu.Y = 21
	menu.Height = 1
	menu.Width = len(menu.Text)

	ui.Render(menu)
}

func (cui *ConsoleUI) drawUpdateAvailable() {
	updateLabel := ui.NewParagraph("(UPDATE AVAILABLE)")
	updateLabel.TextFgColor = ui.StringToAttribute("green")
	updateLabel.Border = false
	updateLabel.X = ui.TermWidth()/2 - 34
	updateLabel.Y = 0
	updateLabel.Width = 19
	updateLabel.Height = 1

	ui.Render(updateLabel)
}

func (cui *ConsoleUI) drawMOTD() {
	motdLabel := ui.NewParagraph(cui.data.MOTD)
	motdLabel.X = ui.TermWidth()/2 - len(cui.data.MOTD)/2
	motdLabel.Border = false
	motdLabel.Y = 12
	motdLabel.Height = 1
	motdLabel.Width = len(cui.data.MOTD) + 1

	ui.Render(motdLabel)
}

func (cui *ConsoleUI) drawStatus() error {
	statusLabel := ui.NewParagraph("")
	var statusString string
	switch cui.data.Status {
	case StatusDevilDaggersNotFound:
		statusString = "Devil Daggers not found"
		statusLabel.TextFgColor = ui.StringToAttribute("red")
	case StatusTitleScreen:
		statusString = "In title screen"
		statusLabel.TextFgColor = ui.StringToAttribute("green")
	case StatusMenu:
		statusString = "In main menu"
		statusLabel.TextFgColor = ui.StringToAttribute("green")
	case StatusLobby:
		statusString = "In dagger lobby"
		statusLabel.TextFgColor = ui.StringToAttribute("green")
	case StatusOwnReplayFromLastRun, StatusOwnReplayFromLeaderboard:
		statusString = "Watching self replay"
		statusLabel.TextFgColor = ui.StringToAttribute("green")
	case StatusOtherReplay:
		statusString = "Watching someone's replay"
		statusLabel.TextFgColor = ui.StringToAttribute("green")
	case StatusPlaying:
		statusString = "Currently playing"
		statusLabel.TextFgColor = ui.StringToAttribute("green")
	case StatusConnecting:
		statusString = "Connecting to Devil Daggers"
		statusLabel.TextFgColor = ui.StringToAttribute("yellow")
	case StatusDead:
		deathType, err := devildaggers.GetDeathTypeString(int(cui.data.DeathType))
		if err != nil {
			return fmt.Errorf("drawStatus: could not get death type string: %w", err)
		}
		statusString = deathType
		statusLabel.TextFgColor = ui.StringToAttribute("red")
	}
	statusLabel.Border = false
	statusLabel.X = ui.TermWidth()/2 - len(statusString)/2 - 19
	statusLabel.Text = "                [[ " + statusString + " ]]                "
	statusLabel.Y = 13
	statusLabel.Height = 1
	statusLabel.Width = 70

	ui.Render(statusLabel)

	return nil
}

func (cui *ConsoleUI) drawOnlineStatus() {
	var onlineLabelText string
	var color ui.Attribute
	switch cui.data.OnlineStatus {
	case OnlineStatusDisconnected:
		color = ui.ColorRed
		onlineLabelText = "[[ Disconnected ]]"
	case OnlineStatusConnecting:
		color = ui.ColorYellow
		onlineLabelText = "[[ Connecting... ]]"
	case OnlineStatusTimedOut:
		color = ui.ColorRed
		onlineLabelText = "  [[ Timed out ]]  "
	case OnlineStatusLoggedIn:
		color = ui.ColorGreen
		onlineLabelText = "  [[ Logged in ]]  "
	case OnlineStatusConnected:
		color = ui.ColorGreen
		onlineLabelText = "  [[ Connected ]]  "
	}
	onlineLabel := ui.NewParagraph(onlineLabelText)
	onlineLabel.TextFgColor = color
	onlineLabel.Border = false
	onlineLabel.X = ui.TermWidth()/2 - len(onlineLabelText)/2
	onlineLabel.Y = 11
	onlineLabel.Height = 1
	onlineLabel.Width = 20

	ui.Render(onlineLabel)
}

func (cui *ConsoleUI) drawRecording() {
	recordingLabel := ui.NewParagraph("")
	if cui.data.Recording {
		recordingLabel.TextFgColor = ui.StringToAttribute("bold, green")
		recordingLabel.Text = "  [[ Recording ]]  "
	} else {
		recordingLabel.TextFgColor = ui.StringToAttribute("red")
		recordingLabel.Text = "[[ Not recording ]]"
	}
	recordingLabel.Border = false
	recordingLabel.X = ui.TermWidth()/2 - len(recordingLabel.Text)/2
	recordingLabel.Y = 14
	recordingLabel.Height = 1
	recordingLabel.Width = 20

	ui.Render(recordingLabel)
}

func (cui *ConsoleUI) drawLeftSideStats() {
	timerString := fmt.Sprintf("In Game Timer: %.4fs", cui.data.Timer)
	daggersHitString := fmt.Sprintf("Daggers Hit: %d", cui.data.DaggersHit)
	daggersFiredString := fmt.Sprintf("Daggers Fired: %d", cui.data.DaggersFired)
	accuracyString := fmt.Sprintf("Accuracy: %.2f%%", cui.data.Accuracy)

	statsLeft := ui.NewParagraph(fmt.Sprintf("%v\n%v\n%v\n%v\n", timerString, daggersHitString, daggersFiredString, accuracyString))
	statsLeft.SetX(ui.TermWidth()/2 - 34)
	statsLeft.SetY(15)
	statsLeft.Border = false
	statsLeft.Width = 34
	statsLeft.Height = 5

	ui.Render(statsLeft)
}

func (cui *ConsoleUI) drawRightSideStats() {
	gemsString := fmt.Sprintf("Gems: %d", cui.data.TotalGems)
	homingString := fmt.Sprintf("Homing Daggers: %d", cui.data.Homing)
	enemiesAliveString := fmt.Sprintf("Enemies Alive: %d", cui.data.EnemiesAlive)
	enemiesKilledString := fmt.Sprintf("Enemies Killed: %d", cui.data.EnemiesKilled)

	statsRight := ui.NewParagraph(fmt.Sprintf("%32v\n%32v\n%32v\n%32v\n", gemsString, homingString, enemiesAliveString, enemiesKilledString))
	statsRight.SetX(ui.TermWidth() / 2)
	statsRight.SetY(15)
	statsRight.Border = false
	statsRight.Width = 34
	statsRight.Height = 5

	ui.Render(statsRight)
}

func (cui *ConsoleUI) drawLastGameLabel() {
	lastGameURL := "None."
	if cui.data.LastGameID != 0 {
		lastGameURL = fmt.Sprintf("%s/games/%d", cui.data.Host, cui.data.LastGameID)
	}

	lastGameLabel := ui.NewParagraph("Last Submission: " + lastGameURL)
	lastGameLabel.SetX(ui.TermWidth()/2 - 34)
	lastGameLabel.SetY(20)
	lastGameLabel.Border = false
	lastGameLabel.Height = 1
	lastGameLabel.Width = 66

	ui.Render(lastGameLabel)
}

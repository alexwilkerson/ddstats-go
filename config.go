package main

type TomlConfig struct {
	SquirrelMode      bool   `toml:"squirrel_mode"`
	GetMOTD           bool   `toml:"get_motd"`
	CheckForUpdates   bool   `toml:"check_for_updates"`
	OfflineMode       bool   `toml:"offline_mode"`
	AutoClipboardGame bool   `toml:"auto_clipboard_game"`
	Host              string `toml:"host"`
	Stream            StreamConfig
	Submit            SubmitConfig
	Discord           DiscordConfig
}

type StreamConfig struct {
	Stats               bool
	ReplayStats         bool `toml:"replay_stats"`
	NonDefaultSpawnsets bool `toml:"non_default_spawnsets"`
}

type SubmitConfig struct {
	Stats               bool
	ReplayStats         bool `toml:"replay_stats"`
	NonDefaultSpawnsets bool `toml:"non_default_spawnsets"`
}

type DiscordConfig struct {
	NotifyAbove1000  bool `toml:"notify_above_1000"`
	NotifyPlayerBest bool `toml:"notify_player_best"`
}

var config = TomlConfig{
	SquirrelMode:      false,
	GetMOTD:           true,
	CheckForUpdates:   true,
	OfflineMode:       false,
	AutoClipboardGame: false,
	Host:              "http://ddstats.com",
	Stream: StreamConfig{
		Stats:               true,
		ReplayStats:         true,
		NonDefaultSpawnsets: true,
	},
	Submit: SubmitConfig{
		Stats:               true,
		ReplayStats:         true,
		NonDefaultSpawnsets: true,
	},
	Discord: DiscordConfig{
		NotifyAbove1000:  true,
		NotifyPlayerBest: true,
	},
}

const defaultConfigFile = `# DDSTATS CONFIGURATION FILE.
# If you mess up this file, press F12 while ddstats.exe is running and the default file will be written.

# "get_motd" retrieve the message of the day from ddstats.com.
# "check_for_updates" check whether there is a new version of ddstats available.
# "offline_mode" if set to true, all networking features will be disabled; the [stream] and [submit] sections will be disabled automatically.
# "auto_clipboard_game" if set to true, the clipboard will be automatically populated with the ddstats url of your last game once it has finished submitting to the server.
# "host" should never be changed. it's here for testing purposes and also so that if i die in a car accident and someone wants to host their own server, they can do so.
get_motd = true
check_for_updates = true
offline_mode = false
auto_clipboard_game = false
# this doesn't really work now because of the update
host = "https://ddstats.com"

# These options are for whether ddstats sends your live game stats to ddstats.com.
# "stats" are your stats in a normal run.
# "replay_stats" are stats while you're watching a replay.
# "non_default_spawnsets" are stats in a run where you are using an alternative survival file.
[stream]
stats = true
replay_stats = true
non_default_spawnsets = true

# These options are for whether ddstats submits your completed games to ddstats.com.
# "stats" are your stats in a normal run.
# "replay_stats" are stats while you're watching a replay.
# "non_default_spawnsets" are stats in a run where you are using an alternative survival file.
[submit]
stats = true
replay_stats = true
non_default_spawnsets = false

# By default, if your game goes above 1000 or if you beat your best time, the ddstats Discord Bot will notify the DevilDaggers.info and DD PALS discord channels. You can disable that feature here.
# "notify_above_1000" notifies when your score goes above 1000 seconds.
# "notify_player_best" notifies when your score goes above your current high score.
[discord]
notify_above_1100 = true
notify_player_best = true`

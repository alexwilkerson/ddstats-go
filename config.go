package main

type tomlConfig struct {
	getMOTD           bool `toml:"get_motd"`
	checkForUpdates   bool `toml:"check_for_updates"`
	offlineMode       bool `toml:"offline_mode"`
	autoClipboardGame bool `toml:"auto_clipboard_game"`
	stream            streamConfig
	submit            submitConfig
	discord           discordConfig
}

type streamConfig struct {
	stats               bool
	replayStats         bool `toml:"replay_stats"`
	nonDefaultSpawnsets bool `toml:"non_default_spawnsets"`
}

type submitConfig struct {
	stats               bool
	replayStats         bool `toml:"replay_stats"`
	nonDefaultSpawnsets bool `toml:"non_default_spawnsets"`
}

type discordConfig struct {
	notifyAbove1000  bool `toml:"notify_above_1000"`
	notifyPlayerBest bool `toml:"notify_player_best"`
}

var config = tomlConfig{
	getMOTD:           true,
	checkForUpdates:   true,
	offlineMode:       false,
	autoClipboardGame: false,
	stream: streamConfig{
		stats:               true,
		replayStats:         true,
		nonDefaultSpawnsets: true,
	},
	submit: submitConfig{
		stats:               true,
		replayStats:         true,
		nonDefaultSpawnsets: true,
	},
	discord: discordConfig{
		notifyAbove1000:  true,
		notifyPlayerBest: true,
	},
}

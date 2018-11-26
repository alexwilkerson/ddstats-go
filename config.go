package main

type config struct {
	classicMode bool `toml:"classic_mode"`
	streamStats bool `toml:"stream_stats"`
}

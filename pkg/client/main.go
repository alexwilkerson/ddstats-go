package client

import (
	"fmt"
	"time"

	"github.com/alexwilkerson/ddstats-go/pkg/config"
	"github.com/alexwilkerson/ddstats-go/pkg/consoleui"
	"github.com/alexwilkerson/ddstats-go/pkg/devildaggers"
	ui "github.com/gizak/termui"
)

const (
	defaultTickRate   = time.Second / 36
	defaultUITickRate = time.Second / 2
)

type Client struct {
	tickRate   time.Duration
	uiTickRate time.Duration
	ui         *consoleui.ConsoleUI
	uiData     *consoleui.Data
	dd         *devildaggers.DevilDaggers
	errChan    chan error
	done       chan struct{}
}

func New() (*Client, error) {
	uiData := consoleui.Data{}

	ui, err := consoleui.New(&uiData)
	if err != nil {
		return nil, fmt.Errorf("New: could not create ui: %w", err)
	}

	dd := devildaggers.New()

	return &Client{
		tickRate:   defaultTickRate,
		uiTickRate: defaultUITickRate,
		ui:         ui,
		uiData:     &uiData,
		dd:         dd,
		errChan:    make(chan error),
		done:       make(chan struct{}),
	}, nil
}

func (c *Client) Run() error {
	defer c.ui.Close()
	defer c.dd.Close()

	go c.run()

	uiEvents := ui.PollEvents()
	select {
	case e := <-uiEvents:
		switch e.ID {
		case "q", "<C-c>", "<f10>":
			close(c.done)
			return nil
		case "<f12>":
			config.WriteDefaultConfigFile()
		case "<MouseLeft>":
			copyGameURLToClipboard()
		}
	case err := <-c.errChan:
		return fmt.Errorf("Run: error returned on error channel: %w", err)
	}

	return nil
}

func (c *Client) run() {
	go c.runUI()
}

func (c *Client) runUI() {
	c.ui.ClearScreen()
	select {
	case <-time.After(c.uiTickRate):
		err := c.ui.DrawScreen()
		if err != nil {
			c.errChan <- fmt.Errorf("runUI: error drawing screen in ui: %w", err)
			return
		}
		c.errChan <- fmt.Errorf("runUI: error drawing screen in ui: %w", err)
	case <-c.done:
		return
	}
}

func copyGameURLToClipboard() {
	// if lastGameURL[:4] == "https" {
	// 	lastGameURLCopyTime = time.Now()
	// 	clipboard.WriteAll(lastGameURL)
	// }
}

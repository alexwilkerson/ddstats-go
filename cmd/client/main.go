package main

import (
	"fmt"
	"os"

	"github.com/alexwilkerson/ddstats-go/pkg/ddstats"
)

const (
	consoleTitle = "ddstats v0.6.0"
)

func main() {
	dd := ddstats.New()

	err := dd.SetConsoleTitle(consoleTitle)
	if err != nil {
		fmt.Printf("main: could not set console title: %v\n", err)
		os.Exit(1)
	}

	err = dd.Connect()
	if err != nil {
		fmt.Println("connection unsuccessful")
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("connection successful")

	err = dd.RefreshData()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

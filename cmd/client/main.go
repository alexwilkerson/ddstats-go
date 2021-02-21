package main

import (
	"fmt"
	"os"

	"github.com/alexwilkerson/ddstats-go/internal/winapi"
)

const (
	consoleTitle = "ddstats v0.6.0"
)

func main() {
	winAPI := winapi.New()

	err := winAPI.SetConsoleTitle(consoleTitle)
	if err != nil {
		fmt.Printf("main: could not set console title: %v\n", err)
		os.Exit(1)
	}

	err = winAPI.Connect()
	if err != nil {
		fmt.Println("connection unsuccessful")
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("connection successful")

	err = winAPI.RefreshDevilDaggersData()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

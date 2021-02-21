package main

import (
	"fmt"
	"os"

	"github.com/alexwilkerson/ddstats-go/internal/winapi"
)

func main() {
	winAPI := winapi.New()

	err := winAPI.Connect()
	if err != nil {
		fmt.Println("connection unsuccessful")
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("connection successful")

	err = winAPI.RefreshDevilDaggersDataBlock()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Scanf("h")
}

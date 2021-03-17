//go:generate goversioninfo

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/alexwilkerson/ddstats-go/pkg/client"
)

const (
	// version must be in "X.X.X" order.
	version        = "0.6.8"
	consoleTitle   = "ddstats v" + version
	v3survivalHash = "569fead87abf4d30fdee4231a6398051"
	grpcAddr       = "172.104.11.117:80"
)

func main() {
	client, err := client.New(version, grpcAddr, v3survivalHash)
	if err != nil {
		if err := logError(err); err != nil {
			log.Fatal(err)
		}
		log.Fatal(err)
	}
	err = client.Run()
	if err != nil {
		if err := logError(err); err != nil {
			log.Fatal(err)
		}
		log.Fatal(err)
	}
}

func logError(inputErr error) error {
	f, err := os.OpenFile("error.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("logError: error opening file: %w", err)
	}
	defer f.Close()

	log.SetOutput(f)
	log.Printf("%v\n", inputErr)
	return nil
}

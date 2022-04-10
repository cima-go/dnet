package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/cima-go/dnet"
	"github.com/cima-go/dnet/rpc"
	"github.com/cima-go/dnet/spec"
)

var mode = flag.String("mode", "", "server|client")

func main() {
	flag.Parse()

	options := dnet.Options{}
	logger, _ := zap.NewDevelopment()

	if *mode == "server" {
		runServer(options, logger)
	} else if *mode == "client" {
		runClient(options, logger)
	} else {
		fmt.Println("Usage: -mode server|client")
	}

}

func runServer(options dnet.Options, logger *zap.Logger) {
	registry := rpc.Init(logger)
	if err := registry.Register("test", func(req []byte) (resp []byte, err error) {
		return req, nil
	}); err != nil {
		log.Fatal(err)
	}

	hw, _ := net.ParseMAC("00:00:5e:00:53:01")

	host := &spec.Identify{
		Hostname:     "test",
		MachineID:    []byte{'T', 'E', 'S', 'T'},
		HardwareAddr: hw,
	}

	server := dnet.Advertise(options, host, registry, logger)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	<-signals
	_ = server.Close()
}

func runClient(options dnet.Options, logger *zap.Logger) {
	disco := dnet.Finder(options, logger)
	peers, err := disco.Peers(time.Second, nil)
	if err != nil {
		log.Fatal(err)
	}

	if len(peers) == 0 {
		log.Println("no peers")
		os.Exit(0)
	}

	for _, peer := range peers {
		log.Println("found peer", peer.Address(), peer.Identify().String())

		ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)

		resp, err := peer.Request(ctx, "test", []byte{0, 0, 1})
		if err != nil {
			log.Println("peer request failed:", err)
		} else {
			log.Println("peer got response:", resp)
		}

		cancel()
	}
}

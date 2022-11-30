package main

import (
	"encoding/hex"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/vulpemventures/go-elements/address"
	"github.com/vulpemventures/ocean/internal/core/domain"
	"github.com/vulpemventures/ocean/internal/core/ports"
	electrum_scanner "github.com/vulpemventures/ocean/internal/infrastructure/blockchain-scanner/electrum"
)

const (
	ADDR = "el1qqtjr65udldf6haazfqlyzrhg09m4c7x6rs40gz836zp9kguvs4zec92nnhqalp0032j9x2sp9v0ejp4g7t5lzng442f355vkh"
)

var bcScanner ports.BlockchainScanner

func startScanner() {
	log.SetLevel(log.DebugLevel)
	var err error
	bcScanner, err = electrum_scanner.NewService(electrum_scanner.ServiceArgs{
		Addr: "tcp://localhost:50001",
	})
	if err != nil {
		log.WithError(err).Fatal("failed to initialize esplora bc scanner")
	}

	bcScanner.Start()
}

func main() {
	startScanner()

	log.RegisterExitHandler(bcScanner.Stop)

	script, _ := address.ToOutputScript(ADDR)
	bcScanner.WatchForAccount("test", 0, []domain.AddressInfo{
		{
			Address:     ADDR,
			Script:      b2h(script),
			BlindingKey: h2b("69259d161aff70409d97d44c29804e46e1153fcf7407d8737b7a79b8b86d2dc9"),
		},
		// {
		// 	Address: "AzprxTQHzYRsppYuSMKNGif9QWbWWZMGraYwQujAQZhZY1opnsnRR5oUEN1EysDQoMxDMJWdA1VHkPvh",
		// 	Script:  "a91458d3b52ffdc832b0aa64d86d953a254af06f4de887",
		// },
		// {
		// 	Address: "CTErLFUxBcnr4tpwpmz28CZCbUCTENSQCJZyV6XDuwnn4UwKrcYdke4oYS1haVwpeDEd4SsvfGszYAdp",
		// 	Script:  "76a91478ceb3b0cc5f9ee0ac03ce0dc0b51f91868300b988ac",
		// },
	})

	go func() {
		for tx := range bcScanner.GetTxChannel("test") {
			fmt.Printf("RECEIVED NEW TX %v %s %t\n", tx.Accounts, tx.TxID, tx.IsConfirmed())
		}
	}()

	go func() {
		for utxos := range bcScanner.GetUtxoChannel("test") {
			fmt.Printf("RECEIVED NEW UTXOS\n")
			for _, u := range utxos {
				fmt.Printf("%+v %t %t %s\n", u.Key(), u.IsSpent(), u.IsConfirmed(), b2h(u.Script))
			}
		}
	}()

	go func() {
		t := time.NewTicker(time.Minute)
		for {
			<-t.C
			_, height, err := bcScanner.GetLatestBlock()
			fmt.Println("CHAIN TIP:", height, err)
		}
	}()
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	<-sigChan

	log.Exit(0)
}

func h2b(str string) []byte {
	buf, _ := hex.DecodeString(str)
	return buf
}

func b2h(buf []byte) string {
	return hex.EncodeToString(buf)
}

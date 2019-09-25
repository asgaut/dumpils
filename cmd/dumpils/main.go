package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/asgaut/dumpils/demod"
	"github.com/bemasher/rtltcp"
)

func main() {
	var sdr rtltcp.SDR

	sdr.HandleFlags()

	// Connect to rtl_tcp server.
	if err := sdr.Connect(nil); err != nil {
		log.Fatal(err)
	}
	defer sdr.Close()

	f := 110.1e6
	channelOffset := 200.0e3
	fs := 10.0 * float64(2<<16) // 1310620.0 Hz

	sdr.SetCenterFreq(uint32(f - channelOffset)) // offset tuning
	sdr.SetSampleRate(uint32(fs))
	sdr.SetGain(40)

	in, out := io.Pipe()
	go func() {
		for {
			io.CopyN(out, sdr, 16384)
		}
	}()

	iqRawData := make([]byte, int(fs/10)<<1)

	periodicTick := time.Tick(time.Second)

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Kill, os.Interrupt)

	demodulator := demod.NewDemodulator(channelOffset, fs)

Loop:
	for {
		select {
		case <-sigint:
			break Loop
		case <-periodicTick:
			//break Loop
		default:
			_, err := io.ReadFull(in, iqRawData)
			if err != nil {
				log.Fatal("Error reading samples:", err)
			}
			p, d, s, i := demodulator.Process(iqRawData)
			// Convert DDM in % to µA
			if f > 200e6 {
				d *= 150 / 17.5
			} else {
				d *= 150 / 15.5
			}
			fmt.Printf("RF(db): %5.1f  DDM(uA):%7.2f  SDM(%%): %6.2f  Ident(%%): %6.2f\n", p, d, s, i)
		}
	}

	fmt.Printf("Exiting.")
}

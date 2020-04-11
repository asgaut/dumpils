package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"

	"github.com/asgaut/dumpils/pkg/demod"
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
	fs := 10.0 * float64(1<<17) // 1310720.0 Hz

	sdr.SetCenterFreq(uint32(f - channelOffset)) // offset tuning
	sdr.SetSampleRate(uint32(fs))
	sdr.SetGain(400) // 40 dB

	in, out := io.Pipe()
	go func() {
		for {
			io.CopyN(out, sdr, 16384)
		}
	}()

	iqRawData := make([]byte, int(fs/10)<<1)

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Kill, os.Interrupt)

	demodulator := demod.NewDemodulator(channelOffset, int(fs/10))

	fmt.Printf("RF(dbFS);DDM(uA);SDM(%%);Ident\n")

Loop:
	for {
		select {
		case <-sigint:
			break Loop
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
			fmt.Printf("%.1f;%.3f;%.3f;%.3f\n", p, d, s, i)
		}
	}

	fmt.Printf("Exiting on Ctrl-C.")
}

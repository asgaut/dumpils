package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
)

var dataSource = map[string]string{
	"loc": "",
	"gp":  "",
}

var processors = map[string]*processor{
	"loc": {},
	"gp":  {},
}

func parseCommandLine() {
	var s1, s2 string
	flag.StringVar(&s1, "loc", "", "address and port of rtl_tcp or filename for LOC data")
	flag.StringVar(&s2, "gp", "", "address and port of rtl_tcp or filename for GP data")
	flag.Parse()
	dataSource["loc"] = s1
	dataSource["gp"] = s2
}

func main() {
	log.SetOutput(os.Stdout)

	parseCommandLine()
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	var cancel context.CancelFunc
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		// Block until a signal is received.
		s := <-sigChan
		log.Println("got signal:", s)
		cancel()
		signal.Stop(sigChan)
	}()

	channelOffset := 200.0e3    // offset tuning
	fs := 10.0 * float64(1<<17) // 1310720.0 Hz

	wg := sync.WaitGroup{}

	for key := range processors {
		go func(src string) {
			wg.Add(1)
			defer wg.Done()
			fmt.Printf("Starting processor for %s\n", src)
			err := error(nil)
			if dataSource[src] != "" {
				if strings.ContainsAny(dataSource[src], ":") {
					err = processors[src].sdrProcess(ctx, dataSource[src], fs)
				} else {
					err = processors[src].fileProcess(ctx, dataSource[src], fs)
				}
			} else {
				err = processors[src].simProcess(ctx, fs)
			}
			if err != nil {
				log.Printf("Error in processor '%s': %v\n", src, err)
			}
			cancel()
		}(key)
	}

	ha := httpapi{
		commands:   make(chan interface{}, 1),
		processors: processors,
	}
	webui := "localhost:3344"
	go func() {
		wg.Add(1)
		defer wg.Done()
		fmt.Printf("Go to http://%s to access the web user interface\n", webui)
		if err := ha.ServeAPI(ctx, webui); err != nil {
			fmt.Println("Server error:", err)
			cancel()
		}
	}()

	fmt.Printf("Demodulating inputs and serving data. Press Ctrl-C to exit.\n")

Loop:
	for {
		select {
		case <-ctx.Done():
			break Loop
		case cmd := <-ha.commands:
			if newChannel, ok := cmd.(channelType); ok {
				fLOC := uint32(newChannel.LOC*1e6 - float32(channelOffset))
				fGP := uint32(newChannel.GP*1e6 - float32(channelOffset))
				log.Printf("Setting frequencies (offset=-%f) %d/%d", channelOffset, fLOC, fGP)
				err0 := processors["loc"].setCenterFreq(fLOC)
				err1 := processors["gp"].setCenterFreq(fGP)
				if err0 != nil || err1 != nil {
					log.Printf("Error setting frequencies %d/%d: '%v' '%v'", fLOC, fGP, err0, err1)
				}
			}
		}
	}

	cancel()
	fmt.Println("Waiting for background tasks to terminate.")
	wg.Wait()
	fmt.Println("Exiting.")
}

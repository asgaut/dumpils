package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"github.com/asgaut/dumpils/pkg/demod2"
	"github.com/bemasher/rtltcp"
)

type processor struct {
	mu          sync.Mutex
	demodulator *demod2.Demodulator
	sdr         rtltcp.SDR
	iqRawData   []byte
}

func (p *processor) setCenterFreq(freq uint32) (err error) {
	if p.sdr.TCPConn != nil {
		return p.sdr.SetCenterFreq(freq)
	}
	return nil
}

func (p *processor) sdrProcess(ctx context.Context, address string, fs float64) error {
	// rtl_test reports these gain values for all my dongles:
	// 0.0 0.9 1.4 2.7 3.7 7.7 8.7 12.5 14.4 15.7 16.6 19.7 20.7 22.9 25.4 28.0 29.7 32.8 33.8 36.4 37.2 38.6 40.2 42.1 43.4 43.9 44.5 48.0 49.6
	addr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return err
	}
	if err := p.sdr.Connect(addr); err != nil {
		return err
	}
	defer p.sdr.Close()
	p.sdr.SetSampleRate(uint32(fs))
	p.sdr.SetGain(40) // must set gain to avoid automatic setting

	p.iqRawData = make([]byte, int(fs/10)*2)
	p.demodulator = demod2.NewDemodulator(fs, int(fs/10))

	go func() {
		<-ctx.Done()
		log.Println("Closing SDR connection")
		// this terminates any read operations
		p.sdr.Close()
	}()

	for {
		_, err := io.ReadFull(p.sdr, p.iqRawData)
		if err != nil {
			log.Println("Error reading from SDR", err)
			return err
		}
		p.mu.Lock()
		p.demodulator.Process(p.iqRawData)
		p.mu.Unlock()
	}
}

func (p *processor) fileProcess(ctx context.Context, filename string, fs float64) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	p.iqRawData = make([]byte, int(fs/10)*2)
	p.demodulator = demod2.NewDemodulator(fs, int(fs/10))

	size, _ := file.Seek(0, io.SeekEnd)
	file.Seek(0, io.SeekStart)
	if size < int64(len(p.iqRawData)) {
		return fmt.Errorf("file size must be at least %d bytes", len(p.iqRawData))
	}
	chunks := int(size / int64(len(p.iqRawData)))

	go func() {
		<-ctx.Done()
		log.Println("Closing input file")
		// this terminates any read operations
		file.Close()
	}()

	loopDuration := time.Duration(int64(time.Second) * int64(len(p.iqRawData)) / int64(fs))
	chunksRead := 0
	for {
		_, err := io.ReadFull(file, p.iqRawData)
		if err != nil {
			return nil
		}
		chunksRead++
		if chunksRead == chunks {
			chunksRead = 0
			file.Seek(0, io.SeekStart)
		}
		p.mu.Lock()
		p.demodulator.Process(p.iqRawData)
		p.mu.Unlock()
		select {
		case <-ctx.Done():
			break
		case <-time.After(loopDuration):
		}
	}
}

func (p *processor) simProcess(ctx context.Context, fs float64) error {
	p.iqRawData = make([]byte, int(fs/10)*2)
	p.demodulator = demod2.NewDemodulator(fs, int(fs/10))
	loopDuration := time.Duration(int64(time.Second) * int64(len(p.iqRawData)) / int64(fs))
	for {
		p.mu.Lock()
		p.demodulator.Process(p.iqRawData)
		p.mu.Unlock()
		select {
		case <-ctx.Done():
			break
		case <-time.After(loopDuration):
		}
	}
}

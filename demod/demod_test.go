package demod

import (
	"math"
	"testing"
)

func complex128ToIQRawData(iqData []complex128, iqRawData []byte) {
	clampToByte := func(v float64) byte {
		if v > 255 {
			v = 255
		} else if v < 0 {
			v = 0
		}
		return byte(v)
	}
	for i := range iqData {
		re := (real(iqData[i]) + 1.0) * 128
		im := (imag(iqData[i]) + 1.0) * 128
		iqRawData[i*2] = clampToByte(re)
		iqRawData[i*2+1] = clampToByte(im)
	}
}

func TestDemod1(t *testing.T) {

	channelOffset := 200.0e3
	fs := 10.0 * float64(2<<16) // 1310720.0 Hz
	n := int(fs / 10)
	iqData := make([]complex128, n)
	iqRawData := make([]byte, int(fs/10)<<1)
	nco := newNCO(channelOffset, fs, n)
	demodulator := NewDemodulator(channelOffset, fs)

	T := 1 / fs
	var time float64
	for i := 0; i < n; i++ {
		iqData[i] = complex(0, 0.5+0.1*math.Sin(2*math.Pi*90*time)+0.1*math.Sin(2*math.Pi*150*time))
		time = time + T
	}
	mult(iqData, nco, iqData)
	complex128ToIQRawData(iqData, iqRawData)
	p, d, s, i := demodulator.Process(iqRawData)

	for i := 0; i < n; i++ {
		iqData[i] = complex(0, 0.5+0.1*math.Sin(2*math.Pi*90*time)+0.1*math.Sin(2*math.Pi*150*time))
		time = time + T
	}
	mult(iqData, nco, iqData)
	complex128ToIQRawData(iqData, iqRawData)
	p, d, s, i = demodulator.Process(iqRawData)
	t.Logf("RF:%.1f dBFS; DDM:%.3f%%; SDM:%.3f%%; Ident:%.3f%%\n", p, d, s, i)
}

// Test demod with 0 channel offset
func TestDemod2(t *testing.T) {

	channelOffset := 0.0        //200.0e3
	fs := 10.0 * float64(2<<16) // 1310720.0 Hz
	n := int(fs / 10)
	iqData := make([]complex128, n)
	T := 1 / fs
	var time float64
	for i := 0; i < n; i++ {
		iqData[i] = complex(0.5+0.1*math.Sin(2*math.Pi*90*time)+0.1*math.Sin(2*math.Pi*150*time), 0)
		time = time + T
	}

	iqRawData := make([]byte, int(fs/10)<<1)
	complex128ToIQRawData(iqData, iqRawData)

	demodulator := NewDemodulator(channelOffset, fs)
	p, d, s, i := demodulator.Process(iqRawData)
	t.Logf("RF:%.1f dBFS; DDM:%.3f%%; SDM:%.3f%%; Ident:%.3f%%\n", p, d, s, i)
}

func TestDemod3(t *testing.T) {

	channelOffset := 200.0e3
	fs := 10.0 * float64(1<<17) // 1310720.0 Hz
	n := int(fs / 10)
	iqData := make([]complex128, n)
	iqRawData := make([]byte, int(fs/10)<<1)
	nco := newNCO(channelOffset+1, fs, n) // add a litte frequency error here so we don't get coherent demod
	demodulator := NewDemodulator(channelOffset, fs)

	T := 1 / fs
	var time float64
	for i := 0; i < n; i++ {
		iqData[i] = complex(0, 0.5+0.1*math.Sin(2*math.Pi*90*time)+0.1*math.Sin(2*math.Pi*150*time))
		time = time + T
	}
	mult(iqData, nco, iqData)
	complex128ToIQRawData(iqData, iqRawData)
	p, d, s, i := demodulator.Process(iqRawData)

	// Generate and process next block
	for i := 0; i < n; i++ {
		iqData[i] = complex(0, 0.5+0.1*math.Sin(2*math.Pi*90*time)+0.1*math.Sin(2*math.Pi*150*time))
		time = time + T
	}
	mult(iqData, nco, iqData)
	complex128ToIQRawData(iqData, iqRawData)
	p, d, s, i = demodulator.Process(iqRawData)
	t.Logf("RF:%.1f dBFS; DDM:%.3f%%=%.1fµA; SDM:%.3f%%; Ident:%.3f%%\n", p, d, d*150/15.5, s, i)
}

// go test .\demod  -v
// or, for CPU usage:
// go test -benchmem -run=^$ github.com/asgaut/dumpils/demod -bench ^(BenchmarkDemod)$

var carrier, p, d, s, i float64

func BenchmarkDemod(b *testing.B) {
	channelOffset := 200.0e3
	fs := 10.0 * float64(2<<16) // 1310720.0 Hz
	iqRawData := make([]byte, int(fs/10)<<1)
	for i := range iqRawData {
		iqRawData[i] = 0
	}
	demodulator := NewDemodulator(channelOffset, fs)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		p, d, s, i = demodulator.Process(iqRawData)
	}
}

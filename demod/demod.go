package demod

import (
	"log"
	"math"
	"math/cmplx"

	"github.com/ktye/fft"
)

func iqToComplex128(input []byte, output []complex128) {
	i := 0
	for idx := range output {
		output[idx] = complex(float64(input[i]), float64(input[i+1]))
		output[idx] /= 127.5
		output[idx] -= (1 + 1i)
		i += 2
	}
}

// mult calculates p = f1 * f2. The slices must be
// preallocated and have the same length.
func mult(f1, f2, p []complex128) {
	for idx := range p {
		p[idx] = f1[idx] * f2[idx]
	}
}

func abs(arr []complex128, res []complex128) {
	for idx := range res {
		res[idx] = complex(cmplx.Abs(arr[idx]), 0)
	}
}

func initNCO(f float64, fs float64, n int) []complex128 {
	nco := make([]complex128, n)
	for idx := range nco {
		nco[idx] = complex128(cmplx.Exp(complex(0, 2*math.Pi*f*float64(idx)/fs)))
	}
	return nco
}

// Demodulator contains preallocated buffers and cached data for the demodulator
type Demodulator struct {
	n       int
	input   []byte
	iqData  []complex128
	fftData []complex128
	ncoData []complex128
	fft     fft.FFT
}

// NewDemodulator creates a Demodulator
func NewDemodulator(channelOffset float64, fs float64) *Demodulator {
	n := int(fs / 10)
	f, err := fft.New(n)
	if err != nil {
		log.Fatal("Error init FFT:", err)
	}
	return &Demodulator{
		//input:   make([]byte, n*2),
		iqData:  make([]complex128, n),
		fftData: make([]complex128, n),
		ncoData: initNCO(-channelOffset, fs, n),
		fft:     f,
	}
}

// Process input samples and calculate ILS measurements
func (d *Demodulator) Process(input []byte) (power, ddm, sdm, ident float64) {
	iqToComplex128(input, d.iqData)
	mult(d.iqData, d.ncoData, d.fftData)
	// Add channel lowpass filter here
	abs(d.fftData, d.fftData) // Demodulate the AM signal
	s := d.fft.Transform(d.fftData)
	carrier := cmplx.Abs(s[0])
	mod150 := (cmplx.Abs(s[15]) + cmplx.Abs(s[len(s)-15])) / carrier * 100
	mod90 := (cmplx.Abs(s[9]) + cmplx.Abs(s[len(s)-9])) / carrier * 100
	ddm = (mod150 - mod90) // 150 Hz dominance (DDM > 0): Fly UP/LEFT
	sdm = (mod150 + mod90)
	ident = (cmplx.Abs(s[102]) + cmplx.Abs(s[len(s)-102])) / carrier * 100
	carrier = carrier / float64(len(s))
	power = 20 * math.Log10(carrier) // Carrier power in dB
	return
}

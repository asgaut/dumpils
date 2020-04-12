package demod2

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
// preallocated and have len(f1) <= len(f2) and len(f1) <= len(p)
func mult(f1, f2, p []complex128) {
	for idx := range f1 {
		p[idx] = f1[idx] * f2[idx]
	}
}

// abs sets the real(out) to the absolute value of the input and imag(out) to zero
func abs(in []complex128, out []complex128) {
	for i := range in {
		out[i] = complex(cmplx.Abs(in[i]), 0)
	}
}

// downsample throws away (len(in)/len(out))-1 input samples per output sample
func downsample(in, out []complex128) {
	n := len(in) / len(out)
	for i := range out {
		out[i] = in[i*n]
	}
}

// Demodulator contains preallocated buffers and cached data for the demodulator
type Demodulator struct {
	n          int
	fft1, fft2 fft.FFT
	fs         float64
	Zero       []complex128
	FFT1       []complex128
	IFFT       []complex128
	LF         []complex128
	Envelope   []complex128
	FFT2       []complex128
	Meas       Meas
}

// Meas holds the demodulated data
type Meas struct {
	Mod150 float32 `json:"mod150"`
	Mod90  float32 `json:"mod90"`
	DDM    float32 `json:"ddm"`
	SDM    float32 `json:"sdm"`
	RF     float32 `json:"rf"`
}

// NewDemodulator creates a Demodulator
func NewDemodulator(fs float64, numSamples int) *Demodulator {
	f1, err := fft.New(numSamples)
	if err != nil {
		log.Fatal("Error init FFT1:", err)
	}
	f2, err := fft.New(numSamples / 16)
	if err != nil {
		log.Fatal("Error init FFT2:", err)
	}
	return &Demodulator{
		//iqData:   make([]complex128, numSamples),
		Zero:     make([]complex128, numSamples),
		FFT1:     make([]complex128, numSamples),
		IFFT:     make([]complex128, numSamples),
		LF:       make([]complex128, numSamples/16),
		Envelope: make([]complex128, numSamples/16),
		FFT2:     make([]complex128, numSamples/16),
		fft1:     f1,
		fft2:     f2,
		n:        numSamples,
		fs:       fs,
	}
}

// Spectrum1 returns the amplitude spectrum before bandpass filtering
func (d *Demodulator) Spectrum1() []float32 {
	s := make([]float32, d.n)
	for i := range s {
		s[i] = float32(cmplx.Abs(d.FFT1[i]) / float64(d.n))
	}
	return s
}

// Spectrum2 returns the amplitude spectrum after bandpass filtering
func (d *Demodulator) Spectrum2() []float32 {
	s := make([]float32, len(d.FFT2))
	for i := range s {
		s[i] = float32(cmplx.Abs(d.FFT2[i]) / float64(len(d.FFT2)))
	}
	return s
}

// Process input samples and calculate ILS measurements.
func (d *Demodulator) Process(input []byte) {
	iqToComplex128(input, d.FFT1)

	// Bandpass filter using FFT and inverse FFT
	// https://dsp.stackexchange.com/questions/6220/why-is-it-a-bad-idea-to-filter-by-zeroing-out-fft-bins
	// (our signals of interest are integer periodic in the FFT width)
	d.fft1.Transform(d.FFT1)
	binFreqWidth := d.fs / float64(d.n)
	passLow, passHigh := int((200e3-5e3)/binFreqWidth), int((200e3+5e3)/binFreqWidth)
	copy(d.IFFT[0:passLow], d.Zero[0:passLow])
	copy(d.IFFT[passLow:passHigh], d.FFT1[passLow:passHigh])
	copy(d.IFFT[passHigh:], d.Zero[passHigh:])
	d.fft1.Inverse(d.IFFT)

	// Downsample and demodulate the AM signal
	downsample(d.IFFT, d.LF)
	abs(d.LF, d.Envelope)

	// FFT to calculate the modulation levels of the navigation tones
	copy(d.FFT2, d.Envelope)
	s := d.fft2.Transform(d.FFT2)

	carrier := cmplx.Abs(s[0])
	d.Meas.Mod150 = float32((cmplx.Abs(s[15]) + cmplx.Abs(s[len(s)-15])) / carrier * 100)
	d.Meas.Mod90 = float32((cmplx.Abs(s[9]) + cmplx.Abs(s[len(s)-9])) / carrier * 100)
	d.Meas.DDM = (d.Meas.Mod150 - d.Meas.Mod90) // 150 Hz dominance (DDM > 0): Fly UP/LEFT
	d.Meas.SDM = (d.Meas.Mod150 + d.Meas.Mod90)
	//ident = (cmplx.Abs(s[102]) + cmplx.Abs(s[len(s)-102])) / carrier * 100
	carrier = carrier / float64(len(s))
	d.Meas.RF = float32(20 * math.Log10(carrier)) // Carrier power in dBFS
}

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

// newNCO returns 'n' unity vectors rotating counter-clockwise at frequency 'f', sampled at frequency 'fs'
func newNCO(w float64, n int) []complex128 {
	nco := make([]complex128, n)
	for i := range nco {
		nco[i] = complex128(cmplx.Exp(complex(0, 2*math.Pi*w*float64(i))))
	}
	return nco
}

const history = 3 // must be >= the order of the IIR filter

func lowpass(in, out []complex128) {
	// 3th order Chebychev Type II (IIR) filter, coeffs normalized (a0 = 1)
	// Filter design with scipy: b, a = signal.cheby2(order=3, rs=60, 25e3/(1310720/2), 'lowpass')
	// y[n] = b[0]*x[n] + b[1]*x[n-1] + b[2]*x[n-2] + b[3]*x[n-3] - a[1]*y[n-1] - a[2]*y[n-2] - a[3]*y[n-3]
	b0, b1, b2, b3 := 0.00017744+0i, -0.00017405+0i, -0.00017405+0i, 0.00017744+0i
	a1, a2, a3 := -2.96202704+0i, 2.92477167+0i, -0.96273785+0i
	for n := history; n < len(in); n = n + 1 {
		out[n] = b0*in[n] + b1*in[n-1] + b2*in[n-2] + b3*in[n-3] - a1*out[n-1] - a2*out[n-2] - a3*out[n-3]
	}

	// Copy the last 'history' samples from from the end to the beginning of the 'in' and 'out' buffers.
	// This makes the previous input and output values available in the for loop above.
	src := len(out) - history
	dst := 0
	for src < len(out) {
		out[dst] = out[src]
		in[dst] = in[src]
		src = src + 1
		dst = dst + 1
	}
}

// Demodulator contains preallocated buffers and cached data for the demodulator
type Demodulator struct {
	n             int
	iqData        []complex128
	lpfIn, lpfOut []complex128
	fftData       []complex128
	nco           []complex128
	fft           fft.FFT
}

// NewDemodulator creates a Demodulator
func NewDemodulator(w float64, numSamples int) *Demodulator {
	f, err := fft.New(numSamples)
	if err != nil {
		log.Fatal("Error init FFT:", err)
	}
	return &Demodulator{
		iqData:  make([]complex128, numSamples),
		lpfIn:   make([]complex128, numSamples+history),
		lpfOut:  make([]complex128, numSamples+history),
		fftData: make([]complex128, numSamples),
		nco:     newNCO(-w, numSamples),
		fft:     f,
		n:       numSamples,
	}
}

// Process input samples and calculate ILS measurements.
// The 'input' time period must be equal to 0.1 seconds.
func (d *Demodulator) Process(input []byte) (power, ddm, sdm, ident float64) {
	iqToComplex128(input, d.iqData)
	mult(d.iqData, d.nco, d.lpfIn[history:history+d.n])
	//lowpass(d.lpfIn, d.lpfOut)
	mult(d.iqData, d.nco, d.lpfOut[history:history+d.n])
	// TODO: subsample lpfOut here?
	abs(d.lpfOut[history:history+d.n], d.fftData) // Demodulate the AM signal
	s := d.fft.Transform(d.fftData)
	carrier := cmplx.Abs(s[0])
	mod150 := (cmplx.Abs(s[15]) + cmplx.Abs(s[len(s)-15])) / carrier * 100
	mod90 := (cmplx.Abs(s[9]) + cmplx.Abs(s[len(s)-9])) / carrier * 100
	ddm = (mod150 - mod90) // 150 Hz dominance (DDM > 0): Fly UP/LEFT
	sdm = (mod150 + mod90)
	ident = (cmplx.Abs(s[102]) + cmplx.Abs(s[len(s)-102])) / carrier * 100
	carrier = carrier / float64(len(s))
	power = 20 * math.Log10(carrier) // Carrier power in dBFS
	return
}

// Process200ms
func (d *Demodulator) Process200ms(input []byte) (power, ddm, sdm, ident float64) {
	iqToComplex128(input, d.iqData)
	mult(d.iqData, d.nco, d.lpfIn[history:history+d.n])
	//lowpass(d.lpfIn, d.lpfOut)
	mult(d.iqData, d.nco, d.lpfOut[history:history+d.n])
	// TODO: subsample lpfOut here?
	abs(d.lpfOut[history:history+d.n], d.fftData) // Demodulate the AM signal
	s := d.fft.Transform(d.fftData)
	carrier := cmplx.Abs(s[0])
	mod150 := (cmplx.Abs(s[15*2]) + cmplx.Abs(s[len(s)-15*2])) / carrier * 100
	mod90 := (cmplx.Abs(s[9*2]) + cmplx.Abs(s[len(s)-9*2])) / carrier * 100
	ddm = (mod150 - mod90) // 150 Hz dominance (DDM > 0): Fly UP/LEFT
	sdm = (mod150 + mod90)
	ident = (cmplx.Abs(s[102*2]) + cmplx.Abs(s[len(s)-102*2])) / carrier * 100
	carrier = carrier / float64(len(s))
	power = 20 * math.Log10(carrier) // Carrier power in dBFS
	return
}

// Process30ms of data
func (d *Demodulator) Process30ms(input []byte) (power, ddm, sdm, ident float64) {
	iqToComplex128(input, d.iqData)
	mult(d.iqData, d.nco, d.lpfIn[history:history+d.n])
	lowpass(d.lpfIn, d.lpfOut)
	abs(d.lpfOut[history:history+d.n], d.fftData) // Demodulate the AM signal
	s := d.fft.Transform(d.fftData)
	carrier := cmplx.Abs(s[0])
	mod150 := (cmplx.Abs(s[5]) + cmplx.Abs(s[len(s)-5])) / carrier * 100
	mod90 := (cmplx.Abs(s[3]) + cmplx.Abs(s[len(s)-3])) / carrier * 100
	ddm = (mod150 - mod90) // 150 Hz dominance (DDM > 0): Fly UP/LEFT
	sdm = (mod150 + mod90)
	ident = (cmplx.Abs(s[34]) + cmplx.Abs(s[len(s)-34])) / carrier * 100
	carrier = carrier / float64(len(s))
	power = 20 * math.Log10(carrier) // Carrier power in dBFS
	return
}

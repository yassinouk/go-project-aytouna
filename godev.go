package main

import (
	"fmt"
	"log"
	"math"
	"math/cmplx"
	"math/rand"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"

	"github.com/mjibson/go-dsp/fft"
)

const (
	channelBandwidth = 10e6
	frequency        = 2e9
	txPower          = 20
	noiseFloor       = -90
	snrThreshold     = 10
	numSubcarriers   = 64
)

type LTEChannel struct{}

func NewLTEChannel() *LTEChannel {
	return &LTEChannel{}
}

func GenerateBits(numBits int) []int64 {
	bits := make([]int64, numBits)
	for i := 0; i < numBits; i++ {
		bits[i] = int64(rand.Intn(2))
	}
	return bits
}

func ModulateBPSK(bits []int64) []complex128 {
	BpskSymbols := make([]complex128, len(bits))
	for i, bit := range bits {
		BpskSymbols[i] = complex(float64(2*bit-1), 0)
	}
	return BpskSymbols
}

func (c *LTEChannel) OFDMModulation(symbols []complex128) []complex128 {
	ofdmSymbols := fft.IFFT(symbols)
	return ofdmSymbols
}

func RayleighChannel(signal []complex128) []complex128 {
	channelGain := complex(rand.NormFloat64(), rand.NormFloat64())

	fadedSignal := make([]complex128, len(signal))
	for i, s := range signal {
		fadedSignal[i] = channelGain * s
	}

	return fadedSignal
}

func (c *LTEChannel) AWGN(signal []complex128, snr float64) []complex128 {
	noisePower := math.Pow(10, (noiseFloor-snr)/10)

	noiseReal := make([]float64, len(signal))
	noiseImag := make([]float64, len(signal))
	for i := range noiseReal {
		noiseReal[i] = rand.NormFloat64() * noisePower / 2
		noiseImag[i] = rand.NormFloat64() * noisePower / 2
	}

	noisySignal := make([]complex128, len(signal))
	for i, s := range signal {
		noisySignal[i] = s + complex(noiseReal[i], noiseImag[i])
	}

	return noisySignal
}

func TransmitSignal(signal []int64, snr float64, lteChannel *LTEChannel) []complex128 {
	BpskSymbols := ModulateBPSK(signal)
	ofdmSymbols := lteChannel.OFDMModulation(BpskSymbols)
	fadedSignal := RayleighChannel(ofdmSymbols)
	receivedSignal := lteChannel.AWGN(fadedSignal, snr)

	return receivedSignal
}

func OFDMDemodulation(receivedSignal []complex128) []complex128 {
	demodulatedSymbols := fft.FFT(receivedSignal)
	return demodulatedSymbols
}

func DemodulateBPSK(receivedSignal []complex128) []int {
	decodedBits := make([]int, len(receivedSignal))
	for i, s := range receivedSignal {
		decodedBits[i] = int(real(s) + 0.5)
	}
	return decodedBits
}

func RunSimulation(numBits int, snr float64) ([]int64, []complex128) {
	lteChannel := NewLTEChannel()

	bits := GenerateBits(numBits)
	fmt.Println("Generated Bits:", bits)

	transmittedBits := bits
	receivedBits := TransmitSignal(bits, snr, lteChannel)

	fmt.Println("Transmitted Bits:\n", transmittedBits)
	fmt.Println("Received Bits:\n", receivedBits)

	return bits, receivedBits
}

func CalculateNorm(c complex128) float64 {
	return cmplx.Abs(c)
}

func plotComplexNorm(points []complex128, title, filename string) {
	p := plot.New()
	p.Title.Text = title
	p.X.Label.Text = "Point Index"
	p.Y.Scale = plot.LogScale{}
	p.Y.Tick.Marker = plot.LogTicks{}
	p.Y.Label.Text = "Magnitude (log scale)"

	bitPoints := make(plotter.XYs, len(points))
	for i := 0; i < len(points); i++ {
		bitPoints[i].X = float64(i)
		bitPoints[i].Y = CalculateNorm(points[i])
	}

	bitLine, err := plotter.NewLine(bitPoints)
	if err != nil {
		log.Panic(err)
	}
	bitLine.LineStyle.Width = vg.Points(1)
	colorIndex := 0
	bitLine.Color = plotutil.Color(colorIndex)

	p.Add(bitLine)
	p.Legend.Add(fmt.Sprintf("Line color %d", colorIndex), bitLine)

	if err := p.Save(6*vg.Inch, 6*vg.Inch, filename); err != nil {
		log.Panic(err)
	}
}

func main() {
	numBits := 10000
	snr := 20.0
	_, receivedBits := RunSimulation(numBits, snr)
	plotComplexNorm(receivedBits, "Received Bit Magnitudes", "magnitudes_plot_received.png")
}

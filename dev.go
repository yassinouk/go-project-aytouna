/* package main

import (
	"fmt"
	"log"
	"math"
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

// LTEChannel struct represents the LTE channel parameters
type LTEChannel struct{}

// NewLTEChannel creates a new LTEChannel instance with default values
func NewLTEChannel() *LTEChannel {
	return &LTEChannel{}
}

// GenerateBits generates random bits using a Bernoulli distribution
func GenerateBits(numBits int) []int64 {
	bits := make([]int64, numBits)
	for i := 0; i < numBits; i++ {
		bits[i] = int64(rand.Intn(2))
	}
	return bits
}

// ModulateBPSK modulates bits using Binary Phase Shift Keying (BPSK)
func ModulateBPSK(bits []int64) []complex128 {
	BpskSymbols := make([]complex128, len(bits))
	for i, bit := range bits {
		BpskSymbols[i] = complex(float64(2*bit-1), 0)
	}
	return BpskSymbols
}

func (c *LTEChannel) OFDMModulation(symbols []complex128) []complex128 {
	// Perform IFFT using the provided IFFT function and specify the number of subcarriers
	ofdmSymbols := fft.IFFT(symbols)
	// Truncate or zero-pad if necessary based on the desired number of subcarriers
	// ofdmSymbols = ofdmSymbols[:numSubcarriers]

	return ofdmSymbols
}

// RayleighChannel simulates a Rayleigh fading channel
func RayleighChannel(signal []complex128) []complex128 {
	// Generate a complex channel gain with Rayleigh distribution
	channelGain := complex(rand.NormFloat64(), rand.NormFloat64())

	// Multiply the signal by the channel gain to simulate fading
	fadedSignal := make([]complex128, len(signal))
	for i, s := range signal {
		fadedSignal[i] = channelGain * s
	}

	return fadedSignal
}

func (c *LTEChannel) AWGN(signal []complex128, snr float64) []complex128 {
	// Calculate noise power based on SNR
	noisePower := math.Pow(10, (noiseFloor-snr)/10)

	// Generate Gaussian noise
	noiseReal := make([]float64, len(signal))
	noiseImag := make([]float64, len(signal))
	for i := range noiseReal {
		noiseReal[i] = rand.NormFloat64() * noisePower / 2
		noiseImag[i] = rand.NormFloat64() * noisePower / 2
	}

	// Add noise to the signal
	noisySignal := make([]complex128, len(signal))
	for i, s := range signal {
		noisySignal[i] = s + complex(noiseReal[i], noiseImag[i])
	}

	return noisySignal
}

func TransmitSignal(signal []complex128, snr float64, lteChannel *LTEChannel) []complex128 {
	// Modulate bits using BPSK
	ofdmSymbols := lteChannel.OFDMModulation(signal)

	// Simulate Rayleigh fading channel
	fadedSignal := RayleighChannel(ofdmSymbols)

	// Add AWGN noise to the signal
	receivedSignal := lteChannel.AWGN(fadedSignal, snr)

	return receivedSignal
}

func OFDMDemodulation(receivedSignal []complex128) []complex128 {
	// Perform FFT using the provided FFT function
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

	BpskSymbols := ModulateBPSK(bits)
	fmt.Println("Symbols BPSK:", BpskSymbols)
	ofdmSymbols := lteChannel.OFDMModulation(BpskSymbols)
	fmt.Println("Symbols OFDM:", ofdmSymbols)
	fadedSignal := RayleighChannel(ofdmSymbols)
	fmt.Println("##################################################")
	fmt.Println("Signal in Rayleigh channel:", fadedSignal)
	fmt.Println("##################################################")
	transmittedBits := bits
	receivedBits := TransmitSignal(BpskSymbols, snr, lteChannel)
	transmittedData := transmittedBits
	receivedData := calculateNormArray(receivedBits)

	fmt.Println("Received Bits:\n", receivedData)
	fmt.Println("transmittedBits:\n", transmittedData)
	fmt.Println("lenght of transmittedBits:\n", len(transmittedData))
	fmt.Println("lenght of receivedData:\n", len(receivedData))

	return bits, receivedBits
}
func calculateNorm(z complex128) float64 {
	// return math.Sqrt(math.Pow(real(z), 2) + math.Pow(imag(z), 2))
	return imag(z)
}
func calculateNormArray(z []complex128) []int64 {
	normArray := make([]int64, len(z))
	for i, z := range z {
		normArray[i] = int64(calculateNorm(z))
	}
	return normArray
}

func plotBits(bits []int64, title, filename string) {
	p := plot.New()
	p.Title.Text = title
	p.X.Label.Text = "Bit Index"
	p.Y.Label.Text = "Bit Value"

	bitPoints := make(plotter.XYs, len(bits))
	for i := 0; i < len(bits); i++ {
		bitPoints[i].X = float64(i)
		bitPoints[i].Y = float64(bits[i])
	}

	bitLine, err := plotter.NewLine(bitPoints)
	if err != nil {
		log.Panic(err)
		return
	}
	bitLine.LineStyle.Width = vg.Points(1)
	bitLine.Color = plotutil.Color(0)

	p.Add(bitLine)
	p.Legend.Add(title, bitLine)

	if err := p.Save(4*vg.Inch, 4*vg.Inch, filename); err != nil {
		log.Panic(err)
	}
}

func main() {
	numBits := 100
	snr := 20.0
	transmittedBits, receivedBits := RunSimulation(numBits, snr)

	transmittedData := transmittedBits
	receivedData := calculateNormArray(receivedBits)

	plotBits(transmittedData, "Transmitted Bits", "bits_plot_transmitted.png")
	plotBits(receivedData, "Received Bits", "bits_plot_received.png")
}
*/
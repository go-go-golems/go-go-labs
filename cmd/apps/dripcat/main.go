package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// StreamBehavior interface
type StreamBehavior interface {
	NextBlockSize() int
	Sleep()
}

func createBehavior(behaviorType string) StreamBehavior {
	switch behaviorType {
	case "naive":
		return &NaiveBehavior{
			BlockSize: naiveBlockSize,
			SizeSpray: naiveSizeSpray,
			Interval:  time.Duration(naiveInterval) * time.Millisecond,
			TimeSpray: time.Duration(naiveTimeSpray) * time.Millisecond,
		}
	case "poisson":
		return &PoissonBehavior{
			Rate: float64(poissonRate) / 1000, // Convert to events per second
		}
	case "bursty":
		return &BurstyBehavior{
			NormalBlockSize:    burstyNormalBlockSize,
			BurstBlockSize:     burstyBurstBlockSize,
			NormalInterval:     time.Duration(burstyNormalInterval) * time.Millisecond,
			BurstInterval:      time.Duration(burstyBurstInterval) * time.Millisecond,
			BurstProbability:   float64(burstyBurstProbability) / 100,
			BurstStateDuration: time.Duration(burstyBurstDuration) * time.Millisecond,
		}
	case "onoff":
		return &OnOffBehavior{
			OnDuration:  time.Duration(onOffOnDuration) * time.Millisecond,
			OffDuration: time.Duration(onOffOffDuration) * time.Millisecond,
			BlockSize:   onOffBlockSize,
		}
	case "pareto":
		return &ParetoBehavior{
			Scale: paretoScale,
			Shape: paretoShape,
		}
	case "selfsimilar":
		return &SelfSimilarBehavior{
			HurstParameter: selfSimilarHurst,
			Mean:           selfSimilarMean,
			Variance:       selfSimilarVariance,
		}
	default:
		fmt.Println("Unknown behavior type, using naive behavior")
		return &NaiveBehavior{BlockSize: 1, Interval: time.Second}
	}
}

type Preset struct {
	Name       string
	Type       string
	Parameters map[string]interface{}
}

var presets = []Preset{
	{
		Name: "naive-fast",
		Type: "naive",
		Parameters: map[string]interface{}{
			"naive-block-size": 10,
			"naive-interval":   100,
		},
	},
	{
		Name: "naive-slow",
		Type: "naive",
		Parameters: map[string]interface{}{
			"naive-block-size": 1,
			"naive-interval":   2000,
		},
	},
	{
		Name: "poisson-burst",
		Type: "poisson",
		Parameters: map[string]interface{}{
			"poisson-rate": 5000,
		},
	},
	{
		Name: "very-bursty",
		Type: "bursty",
		Parameters: map[string]interface{}{
			"bursty-normal-block-size": 1,
			"bursty-burst-block-size":  8,
			"bursty-normal-interval":   2000,
			"bursty-burst-interval":    50,
			"bursty-burst-probability": 30,
			"bursty-burst-duration":    600,
		},
	},
	{
		Name: "bursty",
		Type: "bursty",
		Parameters: map[string]interface{}{
			"bursty-normal-block-size": 1,
			"bursty-burst-block-size":  3,
			"bursty-normal-interval":   400,
			"bursty-burst-interval":    50,
			"bursty-burst-probability": 30,
			"bursty-burst-duration":    600,
		},
	},
	{
		Name: "long-onoff",
		Type: "onoff",
		Parameters: map[string]interface{}{
			"onoff-on-duration":  10000,
			"onoff-off-duration": 10000,
			"onoff-block-size":   5,
		},
	},
	{
		Name: "heavy-tail",
		Type: "pareto",
		Parameters: map[string]interface{}{
			"pareto-scale": 1.5,
			"pareto-shape": 1.2,
		},
	},
	{
		Name: "high-selfsimilar",
		Type: "selfsimilar",
		Parameters: map[string]interface{}{
			"selfsimilar-hurst":    0.9,
			"selfsimilar-mean":     2.0,
			"selfsimilar-variance": 1.5,
		},
	},
}

var (
	behaviorType string

	// Naive behavior parameters
	naiveBlockSize int
	naiveSizeSpray int
	naiveInterval  int
	naiveTimeSpray int

	// Poisson behavior parameters
	poissonRate int

	// Bursty behavior parameters
	burstyNormalBlockSize  int
	burstyBurstBlockSize   int
	burstyNormalInterval   int
	burstyBurstInterval    int
	burstyBurstProbability int
	burstyBurstDuration    int

	// On-Off behavior parameters
	onOffOnDuration  int
	onOffOffDuration int
	onOffBlockSize   int

	// Pareto behavior parameters
	paretoScale float64
	paretoShape float64

	// Self-similar behavior parameters
	selfSimilarHurst    float64
	selfSimilarMean     float64
	selfSimilarVariance float64

	presetName  string
	printPreset bool
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "dripcat [file]",
		Short: "Drip out file content with various streaming behaviors",
		Args:  cobra.MaximumNArgs(1),
		Run:   run,
	}

	rootCmd.Flags().StringVar(&behaviorType, "behavior", "naive", "Streaming behavior type (naive, poisson, bursty, onoff, pareto, selfsimilar, mmpp)")

	// Naive behavior flags
	rootCmd.Flags().IntVar(&naiveBlockSize, "naive-block-size", 1, "Block size for naive behavior")
	rootCmd.Flags().IntVar(&naiveSizeSpray, "naive-size-spray", 0, "Size spray for naive behavior")
	rootCmd.Flags().IntVar(&naiveInterval, "naive-interval", 1000, "Interval in milliseconds for naive behavior")
	rootCmd.Flags().IntVar(&naiveTimeSpray, "naive-time-spray", 0, "Time spray in milliseconds for naive behavior")

	// Poisson behavior flags
	rootCmd.Flags().IntVar(&poissonRate, "poisson-rate", 1000, "Rate (events per second * 1000) for Poisson behavior")

	// Bursty behavior flags
	rootCmd.Flags().IntVar(&burstyNormalBlockSize, "bursty-normal-block-size", 1, "Normal block size for bursty behavior")
	rootCmd.Flags().IntVar(&burstyBurstBlockSize, "bursty-burst-block-size", 5, "Burst block size for bursty behavior")
	rootCmd.Flags().IntVar(&burstyNormalInterval, "bursty-normal-interval", 1000, "Normal interval in milliseconds for bursty behavior")
	rootCmd.Flags().IntVar(&burstyBurstInterval, "bursty-burst-interval", 100, "Burst interval in milliseconds for bursty behavior")
	rootCmd.Flags().IntVar(&burstyBurstProbability, "bursty-burst-probability", 10, "Burst probability (0-100) for bursty behavior")
	rootCmd.Flags().IntVar(&burstyBurstDuration, "bursty-burst-duration", 5000, "Burst duration in milliseconds for bursty behavior")

	// On-Off behavior flags
	rootCmd.Flags().IntVar(&onOffOnDuration, "onoff-on-duration", 5000, "On duration in milliseconds for On-Off behavior")
	rootCmd.Flags().IntVar(&onOffOffDuration, "onoff-off-duration", 5000, "Off duration in milliseconds for On-Off behavior")
	rootCmd.Flags().IntVar(&onOffBlockSize, "onoff-block-size", 1, "Block size for On-Off behavior")

	// Pareto behavior flags
	rootCmd.Flags().Float64Var(&paretoScale, "pareto-scale", 1.0, "Scale parameter for Pareto behavior")
	rootCmd.Flags().Float64Var(&paretoShape, "pareto-shape", 2.0, "Shape parameter for Pareto behavior")

	// Self-similar behavior flags
	rootCmd.Flags().Float64Var(&selfSimilarHurst, "selfsimilar-hurst", 0.7, "Hurst parameter for Self-similar behavior")
	rootCmd.Flags().Float64Var(&selfSimilarMean, "selfsimilar-mean", 1.0, "Mean for Self-similar behavior")
	rootCmd.Flags().Float64Var(&selfSimilarVariance, "selfsimilar-variance", 1.0, "Variance for Self-similar behavior")

	rootCmd.Flags().StringVar(&presetName, "preset", "", "Use a predefined preset")
	rootCmd.Flags().BoolVar(&printPreset, "print-preset", false, "Print available presets and exit")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) {
	if printPreset {
		printPresets()
		return
	}

	if presetName != "" {
		applyPreset(cmd, presetName)
	}

	behavior := createBehavior(behaviorType)

	var reader io.Reader
	if len(args) == 0 || args[0] == "-" {
		reader = os.Stdin
	} else {
		file, err := os.Open(args[0])
		if err != nil {
			fmt.Println("Error opening file:", err)
			os.Exit(1)
		}
		defer file.Close()
		reader = file
	}

	dripContent(reader, behavior)
}

func printPresets() {
	for _, preset := range presets {
		fmt.Printf("Preset: %s\n", preset.Name)
		fmt.Printf("  Type: %s\n", preset.Type)
		fmt.Printf("  Parameters:\n")
		for key, value := range preset.Parameters {
			fmt.Printf("    %s: %v\n", key, value)
		}
		fmt.Println()
	}
}

func applyPreset(cmd *cobra.Command, name string) {
	for _, preset := range presets {
		if strings.EqualFold(preset.Name, name) {
			behaviorType = preset.Type
			for key, value := range preset.Parameters {
				flag := cmd.Flag(key)
				if flag != nil {
					switch v := value.(type) {
					case int:
						_ = flag.Value.Set(fmt.Sprintf("%d", v))
					case float64:
						_ = flag.Value.Set(fmt.Sprintf("%f", v))
					case string:
						_ = flag.Value.Set(v)
					}
				}
			}
			return
		}
	}
	fmt.Printf("Preset '%s' not found. Using default behavior.\n", name)
}

func dripContent(reader io.Reader, behavior StreamBehavior) {
	scanner := bufio.NewScanner(reader)
	buffer := make([]string, 0)

	for scanner.Scan() {
		buffer = append(buffer, scanner.Text())
		if len(buffer) >= behavior.NextBlockSize() {
			printBuffer(buffer)
			buffer = buffer[:0]
			behavior.Sleep()
		}
	}

	if len(buffer) > 0 {
		printBuffer(buffer)
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading input:", err)
	}
}

func printBuffer(buffer []string) {
	for _, line := range buffer {
		fmt.Println(line)
	}
}

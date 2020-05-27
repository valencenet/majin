package generate

import (
	"log"
	"math"
	"math/rand"
	"time"

	vegeta "github.com/tsenart/vegeta/lib"
	"gonum.org/v1/gonum/stat/distuv"
)

// Attacker interface for load attacks.
type Attacker interface {
	Attack(tr vegeta.Targeter, r vegeta.Rate, du time.Duration, name string) <-chan *vegeta.Result
}

// Workload is the parameters we will use to generate a workload simulation.
type Workload struct {
	Load           float64
	Trend          float64
	Period         float64
	Noise          float64
	BurstIndex     float64
	BurstFrequency float64
	BurstDuration  float64
	Duration       float64
}

// Simulation runs a load test simulation based on the workload attack.
func (w *Workload) Simulation(attckr Attacker, targeter vegeta.Targeter) {
	shouldBurst := false
	trend := w.Trend
	period := w.Period
	t, load, amplitude, noise, burst, burstTime := 0.0, 0.0, 0.0, 0.0, 0.0, 0
	noisemaker := distuv.Normal{
		Mu:    float64(w.Noise),
		Sigma: float64(w.Noise),
	}
	for {
		// If our noise flag is 0 than we don't produce noise.
		if w.Noise != 0 {
			noise = math.Abs(noisemaker.Rand())
		} else {
			noise = 0.0
		}

		// Calculate the periodic function (Cosin) with t as the input.
		// Our load is out amplitude. We shift the function vertically by the amplitude + 1.
		amplitude = w.Load / 2
		if period != 0 {
			load = amplitude*math.Cos(math.Pi/(period*2)*t) + (amplitude + 1)
		} else {
			load = w.Load
		}

		if shouldBurst {
			burst = w.BurstIndex
		} else {
			burst = 1.0
		}

		// Determine the per second load we apply to the application.
		rate := vegeta.Rate{Freq: int(math.Round(load+noise+trend) * burst), Per: time.Second}
		log.Printf("Attacking at %f queries per second at %f load, %f trend  and %f noise and a %f burst factor.", math.Ceil(load+noise+trend)*burst, load, trend, noise, burst)
		go func() {
			results := attckr.Attack(targeter, rate, time.Second, "Majin Buu!")
			var metrics vegeta.Metrics
			for res := range results {
				metrics.Add(res)
			}
			metrics.Close()
		}()
		time.Sleep(time.Second)

		// Calculate burst based on how frequent we encounter the Burst Number (1 in BurstFequency).
		if burstTime <= 0 {
			if w.BurstFrequency != 0 && rand.Intn(int(w.BurstFrequency)) == int(w.BurstFrequency-1) {
				shouldBurst = true
				burstTime = int(w.BurstDuration)
			} else {
				shouldBurst = false
			}
		} else {
			burstTime--
		}

		// Increase trend by its basic rate.
		// Increase t by one.
		trend += w.Trend
		t++

		if t == w.Duration {
			break
		}
	}
}

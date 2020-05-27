package generate_test

import (
	"reflect"
	"testing"
	"time"

	vegeta "github.com/tsenart/vegeta/lib"

	"github.com/valencenet/majin/internal/generate"
)

func TestWorkload_Simulation(t *testing.T) {
	type fields struct {
		Load           float64
		Trend          float64
		Period         float64
		Noise          float64
		BurstIndex     float64
		BurstFrequency float64
		BurstDuration  float64
		Duration       float64
	}
	tests := []struct {
		name   string
		fields fields
		want   []int
	}{
		{
			"flat workload is constant",
			fields{
				Load:           10,
				Trend:          0,
				Period:         0,
				Noise:          0,
				BurstIndex:     0,
				BurstFrequency: 0,
				BurstDuration:  0,
				Duration:       10,
			},
			[]int{10, 10, 10, 10, 10, 10, 10, 10, 10, 10},
		},
		{
			"seasonal workload is seasonal",
			fields{
				Load:           10,
				Trend:          0,
				Period:         2,
				Noise:          0,
				BurstIndex:     0,
				BurstFrequency: 0,
				BurstDuration:  0,
				Duration:       9,
			},
			[]int{11, 10, 6, 2, 1, 2, 6, 10, 11},
		},
		{
			"trendy workload is trendy",
			fields{
				Load:           10,
				Trend:          1,
				Period:         0,
				Noise:          0,
				BurstIndex:     0,
				BurstFrequency: 0,
				BurstDuration:  0,
				Duration:       10,
			},
			[]int{11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
		},
		{
			"noisey workload is noisey",
			fields{
				Load:           100,
				Trend:          0,
				Period:         0,
				Noise:          50,
				BurstIndex:     0,
				BurstFrequency: 0,
				BurstDuration:  0,
				Duration:       10,
			},
			[]int{180, 153, 145, 117, 109, 153, 117, 144, 180, 169},
		},
		{
			"bursty workload is bursty",
			fields{
				Load:           10,
				Trend:          0,
				Period:         0,
				Noise:          0,
				BurstIndex:     5,
				BurstFrequency: 2,
				BurstDuration:  5,
				Duration:       10,
			},
			[]int{10, 50, 50, 50, 50, 50, 50, 50, 50, 50},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &generate.Workload{
				Load:           tt.fields.Load,
				Trend:          tt.fields.Trend,
				Period:         tt.fields.Period,
				Noise:          tt.fields.Noise,
				BurstIndex:     tt.fields.BurstIndex,
				BurstFrequency: tt.fields.BurstFrequency,
				BurstDuration:  tt.fields.BurstDuration,
				Duration:       tt.fields.Duration,
			}
			trgtr := vegeta.NewStaticTargeter()
			attckr := &dummyAttacker{}
			w.Simulation(attckr, trgtr)
			if !reflect.DeepEqual(attckr.attackSignature, tt.want) {
				t.Errorf("w.Simulation() failed: \nexpected %v attack signature\n got %v attack signature", tt.want, attckr.attackSignature)
			}
			attckr.flush()
		})
	}
}

type dummyAttacker struct {
	attackSignature []int
}

func (d *dummyAttacker) flush() {
	d.attackSignature = []int{}
}

func (d *dummyAttacker) Attack(tr vegeta.Targeter, r vegeta.Rate, du time.Duration, name string) <-chan *vegeta.Result {
	d.attackSignature = append(d.attackSignature, r.Freq)
	res := make(chan *vegeta.Result)
	close(res)
	return res
}

func (d *dummyAttacker) Stop() {}

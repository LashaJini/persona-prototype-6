package models

import (
	"reflect"
	"testing"
	"time"

	"github.com/wholesome-ghoul/persona-prototype-6/constants"
)

func TestCalculateProbs(t *testing.T) {
	fixtures := []struct {
		desc          string
		input         *WeeklyActivity
		expectedProbs WeekdaysFloat
		startEpoch    int64
		endEpoch      int64
	}{
		{
			desc:          "does nothing when start epoch and end epoch difference is less than threshold",
			input:         &WeeklyActivity{},
			expectedProbs: [DAYS_IN_A_WEEK]float64{},
			startEpoch:    1,
			endEpoch:      constants.EPOCH_DIFF_THRESHOLD,
		},
		{
			desc:          "does nothing if start epoch is greater than end epoch",
			input:         &WeeklyActivity{},
			expectedProbs: [DAYS_IN_A_WEEK]float64{},
			startEpoch:    1,
			endEpoch:      0,
		},
		{
			desc:          "zero probabilities don't change when there is no activity",
			input:         &WeeklyActivity{},
			expectedProbs: [DAYS_IN_A_WEEK]float64{},
			startEpoch:    time.Now().AddDate(0, 0, -7).UTC().Unix(),
			endEpoch:      time.Now().UTC().Unix(),
		},
		{
			desc: "calculates new probabilities #1",
			input: &WeeklyActivity{
				Sun2MonCounts: [DAYS_IN_A_WEEK]int{1, 1, 1, 1, 1, 1, 1},
			},
			expectedProbs: [DAYS_IN_A_WEEK]float64{1, 1, 1, 1, 1, 1, 1},
			startEpoch:    time.Now().AddDate(0, 0, -7).UTC().Unix(),
			endEpoch:      time.Now().UTC().Unix(),
		},
		{
			desc: "calculates new probabilities #2",
			input: &WeeklyActivity{
				Sun2MonCounts: [DAYS_IN_A_WEEK]int{0, 1, 0, 0, 0, 0, 0},
			},
			expectedProbs: [DAYS_IN_A_WEEK]float64{0, 1, 0, 0, 0, 0, 0},
			startEpoch:    time.Now().AddDate(0, 0, -7).UTC().Unix(),
			endEpoch:      time.Now().UTC().Unix(),
		},
		{
			desc: "calculates new probabilities #3",
			input: &WeeklyActivity{
				Sun2MonCounts: [DAYS_IN_A_WEEK]int{1, 1, 1, 1, 1, 1, 1},
			},
			expectedProbs: [DAYS_IN_A_WEEK]float64{1, 1, 1, 1, 1, 1, 1},
			startEpoch:    time.Now().AddDate(0, 0, -7).UTC().Unix(),
			endEpoch:      time.Now().UTC().Unix(),
		},
		{
			desc: "calculates new probabilities #4",
			input: &WeeklyActivity{
				Sun2MonCounts: [DAYS_IN_A_WEEK]int{4, 4, 4, 4, 4, 4, 4},
			},
			expectedProbs: [DAYS_IN_A_WEEK]float64{0.8, 0.8, 0.8, 0.8, 0.8, 0.8, 0.8},
			startEpoch:    time.Now().AddDate(0, 0, -35).UTC().Unix(),
			endEpoch:      time.Now().UTC().Unix(),
		},
	}

	for _, fixture := range fixtures {
		fixture.input.CalculateProbs(fixture.startEpoch, fixture.endEpoch)
		equal := reflect.DeepEqual(fixture.input.Sun2MonProbs, fixture.expectedProbs)
		if !equal {
			t.Errorf("%s\ngot:\n%+v\nexpected:\n%+v", fixture.desc, fixture.input.Sun2MonProbs, fixture.expectedProbs)
		}
	}
}

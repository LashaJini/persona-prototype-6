package main

import (
	"context"
	"time"
)

var yes = struct{}{}

func requestTicker(
	ctx context.Context,
	canMakeRequest chan<- struct{},
	intervalChange <-chan time.Duration,
	duration time.Duration,
) {
	currentInterval := duration
	timer := time.NewTicker(currentInterval)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			canMakeRequest <- yes
		case newInterval := <-intervalChange:
			if newInterval != currentInterval {
				timer.Stop()
				currentInterval = newInterval
				timer = time.NewTicker(currentInterval)
			}
		}
	}
}

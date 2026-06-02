// Package race contains a test with an unsynchronised write.
// Run with -race to trigger the data race detector and see prism's RACE card.
//
//	go test -race -json ./fixture/race/... | prism
package race

import (
	"sync"
	"testing"
)

func TestDataRace(t *testing.T) {
	var counter int
	var wg sync.WaitGroup
	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			counter++ // unsynchronised: DATA RACE when run with -race
		}()
	}
	wg.Wait()
	_ = counter
}

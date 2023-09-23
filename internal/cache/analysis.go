package cache

import (
	"sort"
	"sync"
	"time"

	"golang.org/x/tools/go/analysis"
)

var (
	analyzerRunTimesMu sync.Mutex
	analyzerRunTimes   = make(map[*analysis.Analyzer]time.Duration)
)

type LabelDuration struct {
	Label    string
	Duration time.Duration
}

// AnalyzerTimes returns the accumulated time spent in each Analyzer's
// Run function since process start, in descending order.
func AnalyzerRunTimes() []LabelDuration {
	analyzerRunTimesMu.Lock()
	defer analyzerRunTimesMu.Unlock()

	slice := make([]LabelDuration, 0, len(analyzerRunTimes))
	for a, t := range analyzerRunTimes {
		slice = append(slice, LabelDuration{Label: a.Name, Duration: t})
	}
	sort.Slice(slice, func(i, j int) bool {
		return slice[i].Duration > slice[j].Duration
	})
	return slice
}

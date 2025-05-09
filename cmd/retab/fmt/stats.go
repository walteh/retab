package fmt

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/rs/zerolog"
	"github.com/samber/oops"
)

// does oops have a hook that will let me add the name of the package to every error?

// could we generate specific error codes for things to check?

func applyValueToContext[T any](ctx context.Context, key string, value T) context.Context {
	ctx = oops.WithBuilder(ctx, oops.WithContext(ctx).With(key, value))
	return zerolog.Ctx(ctx).With().Interface(key, value).Logger().WithContext(ctx)
}

func trackStats(ctx context.Context) (context.Context, func(ctx context.Context)) {

	memoryStart := runtime.MemStats{}
	runtime.ReadMemStats(&memoryStart)

	// Track goroutine count
	goroutinesStart := runtime.NumGoroutine()

	// Track GC stats
	gcStart := memoryStart.NumGC

	ctx = zerolog.Ctx(ctx).With().Str("stats", "true").Logger().WithContext(ctx)

	start := time.Now()

	return ctx, func(ctx context.Context) {
		duration := time.Since(start)

		memoryEnd := runtime.MemStats{}
		runtime.ReadMemStats(&memoryEnd)

		// Get final goroutine count
		goroutinesEnd := runtime.NumGoroutine()

		// Calculate additional metrics
		memoryUsage := memoryEnd.TotalAlloc - memoryStart.TotalAlloc
		gcRuns := memoryEnd.NumGC - gcStart

		zerolog.Ctx(ctx).Info().
			Str("go_arch", runtime.GOARCH).
			Str("go_os", runtime.GOOS).
			Str("duration", duration.String()).
			Uint64("memory_usage_bytes", memoryUsage).
			Str("memory_usage_human", humanizeBytes(memoryUsage)).
			Int("goroutines", goroutinesEnd-goroutinesStart).
			Uint32("gc_runs", gcRuns).
			Float64("gc_pause_total_ms", float64(memoryEnd.PauseTotalNs-memoryStart.PauseTotalNs)/1000000).
			Msg("fmt completed")
	}
}

// Helper function to make byte sizes human-readable
func humanizeBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

package util

import (
	"os"
	"runtime/pprof"
)

func InitializeCPUProfiling(path string) {
	if path != "" {
		cpuProfile, err := os.Create(path)
		FailOnError(err)
		err = pprof.StartCPUProfile(cpuProfile)
		FailOnError(err)
		OnExit(pprof.StopCPUProfile)
	}
}

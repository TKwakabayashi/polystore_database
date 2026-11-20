package runner

// pprof.go collects Go profiles of a measured section: <dir>/<label>_<kind>.prof for cpu, heap, allocs,
// block and mutex (block and mutex at rate 1).

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
)

// profileSession is one pprof collection; Stop ends the CPU profile and dumps the others.
type profileSession struct {
	dir, label string
	cpu        *os.File
}

// startProfile begins collecting into dir. Only one CPU profile can run per process, so do not nest it.
func startProfile(dir, label string) *profileSession {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		fmt.Printf("profile: %v\n", err)
		return nil
	}
	s := &profileSession{dir: dir, label: label}
	s.cpu, _ = os.Create(s.path("cpu"))
	if err := pprof.StartCPUProfile(s.cpu); err != nil {
		fmt.Printf("profile: %v\n", err)
	}
	runtime.SetBlockProfileRate(1) // 1ns records every blocking event
	runtime.SetMutexProfileFraction(1)
	return s
}

func (s *profileSession) path(kind string) string {
	return filepath.Join(s.dir, fmt.Sprintf("%s_%s.prof", s.label, kind))
}

// Stop ends the CPU profile and dumps heap, allocs, block and mutex. Safe on nil.
func (s *profileSession) Stop() {
	if s == nil {
		return
	}
	pprof.StopCPUProfile()
	s.cpu.Close()
	runtime.SetBlockProfileRate(0)
	runtime.SetMutexProfileFraction(0)
	dump := func(kind string) {
		f, err := os.Create(s.path(kind))
		if err != nil {
			return
		}
		defer f.Close()
		_ = pprof.Lookup(kind).WriteTo(f, 0)
	}
	runtime.GC() // exact live heap
	dump("heap")
	dump("allocs")
	dump("block")
	dump("mutex")
}

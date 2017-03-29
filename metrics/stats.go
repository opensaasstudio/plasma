package metrics

import (
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

type Stats struct {
	Time int64 `json:"time"`
	// runtime
	GoVersion    string `json:"go_version"`
	GoOs         string `json:"go_os"`
	GoArch       string `json:"go_arch"`
	CpuNum       int    `json:"cpu_num"`
	GoroutineNum int    `json:"goroutine_num"`
	Gomaxprocs   int    `json:"gomaxprocs"`
	CgoCallNum   int64  `json:"cgo_call_num"`
	// memory
	MemoryAlloc      uint64 `json:"memory_alloc"`
	MemoryTotalAlloc uint64 `json:"memory_total_alloc"`
	MemorySys        uint64 `json:"memory_sys"`
	MemoryLookups    uint64 `json:"memory_lookups"`
	MemoryMallocs    uint64 `json:"memory_mallocs"`
	MemoryFrees      uint64 `json:"memory_frees"`
	// stack
	StackInUse uint64 `json:"memory_stack"`
	// heap
	HeapAlloc    uint64 `json:"heap_alloc"`
	HeapSys      uint64 `json:"heap_sys"`
	HeapIdle     uint64 `json:"heap_idle"`
	HeapInuse    uint64 `json:"heap_inuse"`
	HeapReleased uint64 `json:"heap_released"`
	HeapObjects  uint64 `json:"heap_objects"`
	// garbage collection
	GcNext           uint64    `json:"gc_next"`
	GcLast           uint64    `json:"gc_last"`
	GcNum            uint32    `json:"gc_num"`
	GcPerSecond      float64   `json:"gc_per_second"`
	GcPausePerSecond float64   `json:"gc_pause_per_second"`
	GcPause          []float64 `json:"gc_pause"`
	// Connections
	Connections int64 `json:"connections"`
}

type safeTime struct {
	time.Time
	sync.RWMutex
}

func (t *safeTime) get() time.Time {
	t.RLock()
	defer t.RUnlock()
	return t.Time
}

func (t *safeTime) set(tm time.Time) {
	t.Lock()
	defer t.Unlock()
	t.Time = tm
}

var nsInMs float64 = float64(time.Millisecond)

// NOTE: The following three variables need to be changed to atomic
var lastSampleTime safeTime
var lastPauseNs uint64 = 0
var lastNumGc uint32 = 0

var connections int64 = 0

func IncConnection() {
	atomic.AddInt64(&connections, 1)
}

func DecConnection() {
	atomic.AddInt64(&connections, -1)
}

func GetConnection() int64 {
	return atomic.LoadInt64(&connections)
}

func GetStats() *Stats {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	now := time.Now()

	var gcPausePerSecond float64

	if lastPauseNs := atomic.LoadUint64(&lastPauseNs); lastPauseNs > 0 {
		pauseSinceLastSample := mem.PauseTotalNs - lastPauseNs
		gcPausePerSecond = float64(pauseSinceLastSample) / nsInMs
	}

	atomic.SwapUint64(&lastPauseNs, mem.PauseTotalNs)

	var gcPerSecond float64

	lng := atomic.LoadUint32(&lastNumGc)
	countGc := int(mem.NumGC - lng)
	if lng > 0 {
		diff := float64(countGc)
		diffTime := now.Sub(lastSampleTime.get()).Seconds()
		gcPerSecond = diff / diffTime
	}

	if countGc > 256 {
		// lagging GC pause times
		countGc = 256
	}

	gcPause := make([]float64, countGc)

	for i := 0; i < countGc; i++ {
		idx := int((mem.NumGC-uint32(i))+255) % 256
		pause := float64(mem.PauseNs[idx])
		gcPause[i] = pause / nsInMs
	}

	atomic.SwapUint32(&lastNumGc, mem.NumGC)
	lastSampleTime.set(time.Now())

	return &Stats{
		Time:         now.UnixNano(),
		GoVersion:    runtime.Version(),
		GoOs:         runtime.GOOS,
		GoArch:       runtime.GOARCH,
		CpuNum:       runtime.NumCPU(),
		GoroutineNum: runtime.NumGoroutine(),
		Gomaxprocs:   runtime.GOMAXPROCS(0),
		CgoCallNum:   runtime.NumCgoCall(),
		// memory
		MemoryAlloc:      mem.Alloc,
		MemoryTotalAlloc: mem.TotalAlloc,
		MemorySys:        mem.Sys,
		MemoryLookups:    mem.Lookups,
		MemoryMallocs:    mem.Mallocs,
		MemoryFrees:      mem.Frees,
		// stack
		StackInUse: mem.StackInuse,
		// heap
		HeapAlloc:    mem.HeapAlloc,
		HeapSys:      mem.HeapSys,
		HeapIdle:     mem.HeapIdle,
		HeapInuse:    mem.HeapInuse,
		HeapReleased: mem.HeapReleased,
		HeapObjects:  mem.HeapObjects,
		// garbage collection
		GcNext:           mem.NextGC,
		GcLast:           mem.LastGC,
		GcNum:            mem.NumGC,
		GcPerSecond:      gcPerSecond,
		GcPausePerSecond: gcPausePerSecond,
		GcPause:          gcPause,
		// Connections
		Connections: GetConnection(),
	}
}

package memoryrescue

import (
	"sort"
	"sync"
	"sync/atomic"
)

const (
	minimumBitSize          = 6 // 2**6=64 is a CPU cache line size
	steps                   = 20
	minimumSize             = 1 << minimumBitSize
	maximumSize             = 1 << (minimumBitSize + steps - 1)
	calibrateCallsThreshold = 42000
	maximumPercentile       = 0.95
)

type BuffPool struct {
	calls       [steps]uint64
	calibrating int64
	defaultSize uint64
	maxSize     uint64
	pool        sync.Pool
}

var defaultPool BuffPool

type callSize struct {
	calls uint64
	size  uint64
}

type callSizes []callSize

func (ci callSizes) Len() int {
	return len(ci)
}

func (ci callSizes) Less(i, j int) bool {
	return ci[i].calls > ci[j].calls
}

func (ci callSizes) Swap(i, j int) {
	ci[i], ci[j] = ci[j], ci[i]
}

func Get() *Buffer { return defaultPool.Get() }
func (p *BuffPool) Get() *Buffer {
	v := p.pool.Get()

	if v != nil {
		return v.(*Buffer)
	}

	return &Buffer{buff: make([]byte, 0, atomic.LoadUint64(&p.defaultSize))}
}

func Put(bf *Buffer) { defaultPool.Put(bf) }
func (p *BuffPool) Put(bf *Buffer) {
	index := findIndex(len(bf.buff))

	if atomic.AddUint64(&p.calls[index], 1) > calibrateCallsThreshold {
		p.calibrate()
	}

	maxSize := int(atomic.LoadUint64(&p.maxSize))

	if maxSize == 0 || cap(bf.buff) < maxSize {
		bf.Reset()
		p.pool.Put(bf)
	}
}

func (p *BuffPool) calibrate() {
	if !atomic.CompareAndSwapInt64(&p.calibrating, 0, 1) {
		return
	}

	a := make(callSizes, 0, steps)
	var callSum uint64

	for i := uint64(0); i < steps; i++ {
		calls := atomic.SwapUint64(&p.calls[i], 0)
		callSum += calls
		a = append(a, callSize{calls: calls, size: minimumSize << i})
	}
	sort.Sort(a)
	defaultSize := a[0].size
	maximumSize := defaultSize

	maxSum := uint64(float64(callSum) * maximumPercentile)
	callSum = 0

	for i := 0; i < steps; i++ {
		if callSum > maxSum {
			break
		}

		callSum += a[i].calls
		size := a[i].size

		if size > maximumSize {
			maximumSize = size
		}
	}
	atomic.StoreUint64(&p.defaultSize, defaultSize)
	atomic.StoreUint64(&p.maxSize, maximumSize)
	atomic.StoreInt64(&p.calibrating, 0)
}

func findIndex(n int) int {
	n--
	n >>= minimumBitSize
	index := 0

	for n > 0 {
		n >>= 1
		index++
	}

	if index > steps {
		index = steps - 1
	}

	return index
}

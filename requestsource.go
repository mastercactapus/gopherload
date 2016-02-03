package gopherload

import (
	"math/rand"
	"net/http"
	"sort"
	"sync/atomic"
)

// RequestSource implements the NewRequest method for creating new requests
type RequestSource interface {
	NewRequest() (*http.Request, error)
}

// RoundRobinRequestSource is used to supply requests in a round-robin fasion from a slice of RequestSource
type RoundRobinRequestSource struct {
	Sources []RequestSource
	n       int64
}

// NewRequest will create a new request. The source will be chosen in round-robin fasion.
// It will panic if .Sources is empty or nil. NewRequest may be called from multiple goroutines.
func (r *RoundRobinRequestSource) NewRequest() (*http.Request, error) {
	i := atomic.AddInt64(&r.n, 1)
	return r.Sources[(int(i)-1)%len(r.Sources)].NewRequest()
}

// RandomRequestSource is used to supply random requests. Sources is the slice of
// RequestSource to use. Random is an optional function for generating a
// non-negative number in [0,n). If no function is provided "math/rand".Intn will be used.
type RandomRequestSource struct {
	Sources    []RequestSource
	Random     func(n int) int
	cdfWeights []int
	cdfTotal   int
}

func (s *RandomRequestSource) intn(n int) int {
	if n <= 0 {
		panic("invalid range")
	}
	if s.Random != nil {
		return s.Random(n)
	}
	return rand.Intn(n)
}

// NewRequest will create a new request. The source will be chosen at random. It will
// panic if .Sources is empty or nil. NewRequest may be called from multiple goroutines.
func (s *RandomRequestSource) NewRequest() (*http.Request, error) {
	var index int
	if s.cdfWeights != nil {
		index = selectCdf(s.cdfWeights, s.intn(s.cdfTotal))
	} else {
		index = s.intn(len(s.Sources))
	}

	return s.Sources[index].NewRequest()
}

// SetWeights will configure weights to individual Sources. If nil, weighting will be
// disabled. Weights < 1 will disable a source. If non-nil it will panic if len(weights) != len(Sources)
// If .Sources is changed and weights are being used, SetWeights must be called again.
func (s *RandomRequestSource) SetWeights(weights []int) {
	if weights == nil {
		s.cdfWeights = nil
	} else {
		s.cdfWeights, s.cdfTotal = makeCdf(weights)
	}
}

func selectCdf(cdf []int, val int) int {
	return sort.Search(len(cdf), func(i int) bool {
		return cdf[i] > val
	})
}

func makeCdf(weights []int) ([]int, int) {
	cdf := make([]int, len(weights))
	var n int
	for i, w := range weights {
		if w > 0 {
			n += w
		}
		cdf[i] = n
	}
	return cdf, n
}

// SelectableRequestSource is a thread-safe way to switch between sources.
type SelectableRequestSource struct {
	Sources []RequestSource
	n       int64
}

// NewRequest will create a new request from the current source.
func (s *SelectableRequestSource) NewRequest() (*http.Request, error) {
	return s.Sources[int(atomic.LoadInt64(&s.n))].NewRequest()
}

// Index returns the currently selected index
func (s *SelectableRequestSource) Index() int {
	return int(atomic.LoadInt64(&s.n))
}

// SetIndex updates the currently selected index. SetIndex is safe to call from multiple
// goroutines. It will panic if index is < 0 or >= len(sources)
func (s *SelectableRequestSource) SetIndex(index int) {
	if index < 0 || index >= len(s.Sources) {
		panic("index out of range")
	}
	atomic.StoreInt64(&s.n, int64(index))
}

package diameter

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"sync"
	"time"
)

// A SequenceGenerator is used to provide monotonically increasing values for Diameter Message
// hop-by-hop IDs and end-to-end IDs.
type SequenceGenerator struct {
	hbhGenerator *HopByHopIdGenerator
	eteGenerator *EndToEndIdGenerator
}

// NewSequenceGeneratorSet creates a new SequenceGenerator with the hop-by-hop ID seed set to
// a random uint32 value and the end-to-end ID lower 24-bits seeded with a random 24-bit integer
// value.
func NewSequenceGeneratorSet() *SequenceGenerator {
	return &SequenceGenerator{
		NewHopByHopIdGenerator(),
		NewEndToEndIdGenerator(),
	}
}

// NextHopByHopId returns the next hop-by-hop ID in the sequence.  It will be equal to the last
// value supplied (or the seed on the first invocation of this method) plus 1.  If the limit of
// a uint32 is reached, then 0 is returned.  It is safe to call this method in multiple
// coroutines simultaneously.
func (g *SequenceGenerator) NextHopByHopId() uint32 {
	return g.hbhGenerator.Next()
}

// NextEndToEndId return the next end-to-end ID in the sequence.  The high-order 8 bits is
// the low-order 8 bits in the current unix epoch time in seconds.  The low-order 24 bits
// is a value that increments by one on each call, starting with the seed value.  If the
// low-order 24-bits value exceeds the limit of a 24-bit unsigned integer, it wraps to 0 and
// continues incrementing from there.  It is safe to call this method in multiple
// coroutines simultaneously.
func (g *SequenceGenerator) NextEndToEndId() uint32 {
	return g.eteGenerator.Next()
}

// A HopByHopIdGenerator is used to generate monotonically increasing hop-by-hop IDs
// starting with a random seed.
type HopByHopIdGenerator struct {
	mu        sync.Mutex
	nextValue uint32
}

// An EndToEndIdGenerator is used to generate monotonically increasingly end-to-end IDs
// using the method described in RFC6733.  The high-order 8 bits of an ID is the low-eight
// bits of the unix epoch time in seconds.  The low-order 24 bits start with a random 24-bit
// integer value and increments on each call.  If the low-order bits exceeds the limit of
// a 24-bit unsigned integer, it wraps to 0 and continues incrementing from there.
type EndToEndIdGenerator struct {
	mu                      sync.Mutex
	nextValueForLower24Bits uint32
}

// NewHopByHopIdGenerator returns a HopByHopIdGenerator with the initial seed set to a random
// uint32 value.
func NewHopByHopIdGenerator() *HopByHopIdGenerator {
	n, err := rand.Int(rand.Reader, big.NewInt(0xffffffff))
	if err != nil {
		panic(fmt.Errorf("failed to generate random integer: %s", err))
	}

	return &HopByHopIdGenerator{
		nextValue: uint32(n.Uint64()),
	}
}

// Next returns the next ID according to the rules described above.  It is safe to call
// this method in multiple coroutines simultaneously.
func (g *HopByHopIdGenerator) Next() uint32 {
	g.mu.Lock()
	defer g.mu.Unlock()

	n := g.nextValue
	g.nextValue++
	return n
}

// NewEndToEndIdGenerator returns an EndToEndIdGenerator with the low-order 24-bit
// seed set to a random 24-bit unsigned integer.
func NewEndToEndIdGenerator() *EndToEndIdGenerator {
	n, err := rand.Int(rand.Reader, big.NewInt(0xffffffff))
	if err != nil {
		panic(fmt.Errorf("failed to generate random integer: %s", err))
	}

	return &EndToEndIdGenerator{
		nextValueForLower24Bits: uint32(n.Uint64()),
	}
}

// Next returns the next ID according to the rules described above.  It is safe to call
// this method in multiple coroutines simultaneously.
func (g *EndToEndIdGenerator) Next() uint32 {
	now := time.Now().Unix()

	g.mu.Lock()
	n := g.nextValueForLower24Bits
	g.nextValueForLower24Bits++
	g.mu.Unlock()

	return ((uint32(now) & 0xff) << 24) | (n & 0x00ffffff)
}

// GenerateSessionId is used to generate a Session-Id using the mechanism described in
// RFC6733.  Specifically, given an originHost value, it produces
// <originHost>;<time-high>;<time-low>.  "time" here is the number of microseconds since
// the Unix epoch as a uint64.  "high" is the high-order 32 bits of this number and "low"
// is the low-order 32 bits of this number.
func GenerateSessionId(originHost string) string {
	now := uint64(time.Now().UnixMicro())
	return fmt.Sprintf("%s;%d;%d", originHost, uint32(now>>32), uint32(now))
}

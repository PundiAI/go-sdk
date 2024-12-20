package rand

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRandStr(t *testing.T) {
	l := 243
	s := Str(l)
	assert.Len(t, s, l)
}

func TestRandBytes(t *testing.T) {
	l := 243
	b := Bytes(l)
	assert.Len(t, b, l)
}

func TestRandIntn(t *testing.T) {
	n := 243
	for i := 0; i < 100; i++ {
		x := Intn(n)
		assert.Less(t, x, n)
	}
}

// Test to make sure that we never call math.rand().
// We do this by ensuring that outputs are deterministic.
func TestDeterminism(t *testing.T) {
	var firstOutput string

	for i := 0; i < 100; i++ {
		output := testThemAll()
		if i == 0 {
			firstOutput = output
		} else if firstOutput != output {
			t.Errorf("run #%d's output was different from first run.\nfirst: %v\nlast: %v",
				i, firstOutput, output)
		}
	}
}

func testThemAll() string {
	// Such determinism.
	grand.reset(1)

	// Use it.
	out := new(bytes.Buffer)
	perm := Perm(10)
	blob, _ := json.Marshal(perm)
	_, _ = fmt.Fprintf(out, "perm: %s\n", blob)
	_, _ = fmt.Fprintf(out, "randInt: %d\n", Int())
	_, _ = fmt.Fprintf(out, "randUint: %d\n", Uint())
	_, _ = fmt.Fprintf(out, "randIntn: %d\n", Intn(97))
	_, _ = fmt.Fprintf(out, "randInt31: %d\n", Int31())
	_, _ = fmt.Fprintf(out, "randInt32: %d\n", Int32())
	_, _ = fmt.Fprintf(out, "randInt63: %d\n", Int63())
	_, _ = fmt.Fprintf(out, "randInt64: %d\n", Int64())
	_, _ = fmt.Fprintf(out, "randUint32: %d\n", Uint32())
	_, _ = fmt.Fprintf(out, "randUint64: %d\n", Uint64())
	return out.String()
}

func TestRngConcurrencySafety(*testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			_ = Uint64()
			<-time.After(time.Millisecond * time.Duration(Intn(100)))
			_ = Perm(3)
		}()
	}
	wg.Wait()
}

func BenchmarkRandBytes10B(b *testing.B) {
	benchmarkRandBytes(b, 10)
}

func BenchmarkRandBytes100B(b *testing.B) {
	benchmarkRandBytes(b, 100)
}

func BenchmarkRandBytes1KiB(b *testing.B) {
	benchmarkRandBytes(b, 1024)
}

func BenchmarkRandBytes10KiB(b *testing.B) {
	benchmarkRandBytes(b, 10*1024)
}

func BenchmarkRandBytes100KiB(b *testing.B) {
	benchmarkRandBytes(b, 100*1024)
}

func BenchmarkRandBytes1MiB(b *testing.B) {
	benchmarkRandBytes(b, 1024*1024)
}

func benchmarkRandBytes(b *testing.B, n int) {
	b.Helper()
	for i := 0; i < b.N; i++ {
		_ = Bytes(n)
	}
	b.ReportAllocs()
}

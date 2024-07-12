package strmap

import (
	"crypto/rand"
	"runtime"
	"testing"

	"github.com/cloudwego/gopkg/internal/unsafe"
	"github.com/stretchr/testify/require"
)

func randString(m int) string {
	b := make([]byte, m)
	rand.Read(b)
	return string(b)
}

func randStrings(m, n int) []string {
	b := make([]byte, m*n)
	rand.Read(b)
	ret := make([]string, 0, n)
	for i := 0; i < n; i++ {
		s := b[m*i:]
		s = s[:m]
		ret = append(ret, unsafe.ByteSliceToString(s))
	}
	return ret
}

func newStdStrMap(ss []string) map[string]uintptr {
	v := uintptr(1)
	m := make(map[string]uintptr)
	for _, s := range ss {
		_, ok := m[s]
		if !ok {
			m[s] = v
			v++
		}
	}
	return m
}

func TestStrMap(t *testing.T) {
	ss := randStrings(20, 100000)
	m := newStdStrMap(ss)
	sm := New(m)
	require.Equal(t, len(m), sm.Len())
	for i, s := range ss {
		v0 := m[s]
		v1, _ := sm.Get(s)
		require.Equal(t, v0, v1, i)
	}
	for i, s := range randStrings(20, 100000) {
		v0, ok0 := m[s]
		v1, ok1 := sm.Get(s)
		require.Equal(t, ok0, ok1, i)
		require.Equal(t, v0, v1, i)
	}
	m0 := make(map[string]uintptr)
	for i := 0; i < sm.Len(); i++ {
		s, v := sm.Item(i)
		m0[s] = v
	}
	require.Equal(t, m, m0)
}

func TestStrMapString(t *testing.T) {
	ss := []string{"a", "b", "c"}
	m := newStdStrMap(ss)
	sm := New(m)
	t.Log(sm.String())
	t.Log(sm.DebugString())
}

func Benchmark_StrMap(b *testing.B) {
	ss := randStrings(20, 100000)
	m := newStdStrMap(ss)
	sm := New(m)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = sm.Get(ss[i%len(ss)])
	}
}

func Benchmark_StdMap(b *testing.B) {
	ss := randStrings(20, 100000)
	m := newStdStrMap(ss)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m[ss[i%len(ss)]]
	}
}

func Benchmark_StrMap_GC(b *testing.B) {
	ss := randStrings(50, 1000000)
	m := newStdStrMap(ss)
	sm := New(m)
	ss = nil
	m = nil
	runtime.GC()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runtime.GC()
	}
	runtime.KeepAlive(sm)
}

func Benchmark_StdMap_GC(b *testing.B) {
	ss := randStrings(50, 1000000)
	m := newStdStrMap(ss)
	ss = nil
	runtime.GC()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runtime.GC()
	}
	runtime.KeepAlive(m)
}

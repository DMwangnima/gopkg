package strmap

import (
	"time"
	"unsafe"
)

const (
	fnvHashOffset64 = uint64(14695981039346656037) // fnv hash offset64
	fnvHashPrime64  = uint64(1099511628211)
)

var hashseed = fnvHashOffset64

func init() {
	hashseed = hashstr(time.Now().Format(time.RFC3339Nano))
}

func strDataPtr(s string) unsafe.Pointer {
	// XXX: for str or slice, the Data ptr is always the 1st field
	return *(*unsafe.Pointer)(unsafe.Pointer(&s))
}

func hashstr(s string) uint64 {
	// a modified version of fnv hash,
	// it computes 8 bytes per round,
	// and doesn't generate the same result for diff cpu arch,
	// so it's ok for in-memory use

	h := hashseed
	p := strDataPtr(s)

	// 8 byte per round
	i := 0
	for n := len(s) >> 3; i < n; i++ {
		h *= fnvHashPrime64
		h ^= *(*uint64)(unsafe.Add(p, i<<3)) // p[i*8]
	}

	// left 0-7 bytes
	i = i << 3
	for ; i < len(s); i++ {
		h *= fnvHashPrime64
		h ^= uint64(s[i])
	}
	return h
}

var bits2primes = []uint{
	0:  17,         // 1
	1:  17,         // 2
	2:  17,         // 4
	3:  17,         // 8
	4:  17,         // at least 17 for <= 16
	5:  31,         // 32
	6:  61,         // 64
	7:  127,        // 128
	8:  251,        // 256
	9:  509,        // 512
	10: 1021,       // 1024
	11: 2039,       // 2048
	12: 4093,       // 4096
	13: 8191,       // 8192
	14: 16381,      // 16384
	15: 32749,      // 32768
	16: 65521,      // 65536
	17: 131071,     // 131072
	18: 262139,     // 262144
	19: 524287,     // 524288
	20: 1048573,    // 1048576
	21: 2097143,    // 2097152
	22: 4194301,    // 4194304
	23: 8388593,    // 8388608
	24: 16777213,   // 16777216
	25: 33554393,   // 33554432
	26: 67108859,   // 67108864
	27: 134217689,  // 134217728
	28: 268435399,  // 268435456
	29: 536870909,  // 536870912
	30: 1073741789, // 1073741824
}

func calcHashtableSlots(n int) uint {
	// load factor
	n = int(float32(n) / 0.75)

	// count bits to decide which prime number to use
	bits := 0
	for v := uint64(n); v > 0; v = v >> 1 {
		bits++
	}

	// add one more bit,
	// so if n=1500, than returns 2039 instead of 1021
	bits++

	if bits > len(bits2primes) {
		// ???? are you sure we need to hold so many items? ~ 1B items for 30 bits
		return uint(n)
	}
	return bits2primes[bits] // a prime bigger than n
}

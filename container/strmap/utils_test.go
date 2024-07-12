package strmap

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHashStr(t *testing.T) {
	require.Equal(t, hashstr("1234"), hashstr("1234"))
	require.NotEqual(t, hashstr("12345"), hashstr("12346"))
	require.Equal(t, hashstr("12345678"), hashstr("12345678"))
	require.NotEqual(t, hashstr("123456789"), hashstr("123456788"))
}

func BenchmarkHashStr(b *testing.B) {
	strSizes := []int{8, 16, 32, 64, 128, 512}
	ss := make([]string, len(strSizes))
	for i := range ss {
		ss[i] = randString(strSizes[i])
	}
	b.ResetTimer()
	for _, s := range ss {
		b.Run(fmt.Sprintf("size-%d", len(s)), func(b *testing.B) {
			b.SetBytes(int64(len(s)))
			for i := 0; i < b.N; i++ {
				_ = hashstr(s)
			}
		})
	}
}

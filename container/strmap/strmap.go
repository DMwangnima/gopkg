package strmap

import (
	"fmt"
	"sort"
	"strings"

	"github.com/cloudwego/gopkg/internal/unsafe"
)

// StrMap represents GC friendly string map implementation.
// it's readonly after it's created
type StrMap struct {
	data  []byte
	items []mapItem

	hashtable []int
}

type mapItem struct {
	off  int
	sz   int
	slot uint
	v    uintptr
}

// New creates StrMap from map[string]uintptr
// uintptr can be any value and it will be returned by Get.
func New(m map[string]uintptr) *StrMap {
	sz := 0
	for k, _ := range m {
		sz += len(k)
	}
	b := make([]byte, 0, sz)
	items := make([]mapItem, 0, len(m))
	for k, v := range m {
		items = append(items, mapItem{off: len(b), sz: len(k), slot: uint(hashstr(k)), v: v})
		b = append(b, k...)
	}
	ret := &StrMap{data: b, items: items}
	ret.makeHashtable()
	return ret
}

// Len returns the size of map
func (m *StrMap) Len() int {
	return len(m.items)
}

// Item returns the i'th item in map.
// It panics if i is not in the range [0, Len()).
func (m *StrMap) Item(i int) (string, uintptr) {
	e := &m.items[i]
	return unsafe.ByteSliceToString(m.data[e.off : e.off+e.sz]), e.v
}

func (m *StrMap) makeHashtable() {
	slots := calcHashtableSlots(len(m.items))
	m.hashtable = make([]int, slots)

	for i := range m.items {
		m.items[i].slot = m.items[i].slot % slots
	}

	// make sure items with the same slot stored together
	// good for cpu cache
	sort.Slice(m.items, func(i, j int) bool {
		return m.items[i].slot < m.items[j].slot
	})

	for i := 0; i < len(m.hashtable); i++ {
		m.hashtable[i] = -1
	}
	for i := range m.items {
		e := &m.items[i]
		if m.hashtable[e.slot] < 0 {
			// we only need to store the 1st item if hash conflict
			// since they're already stored together
			// will check the next item when Get
			m.hashtable[e.slot] = i
		}
	}
}

// Get ...
func (m *StrMap) Get(s string) (uintptr, bool) {
	slot := uint(hashstr(s)) % uint(len(m.hashtable))
	i := m.hashtable[slot]
	if i < 0 {
		return 0, false
	}
	e := &m.items[i]
	for {
		if string(m.data[e.off:e.off+e.sz]) == s { // double check
			return e.v, true
		}
		i++
		if i >= len(m.items) {
			break
		}
		e = &m.items[i]
		if e.slot != slot { // items sorted by slot
			break
		}
	}
	return 0, false
}

func (m *StrMap) String() string {
	b := &strings.Builder{}
	b.WriteByte('{')
	for i, e := range m.items {
		if i != 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(b, "%q: %x", string(m.data[e.off:e.off+e.sz]), e.v)
	}
	b.WriteByte('}')
	return b.String()
}

func (m *StrMap) DebugString() string {
	b := &strings.Builder{}
	for _, e := range m.items {
		fmt.Fprintf(b, "{off:%d, slot:%x, str:%q, v:%x}\n", e.off, e.slot, string(m.data[e.off:e.off+e.sz]), e.v)
	}
	return b.String()
}

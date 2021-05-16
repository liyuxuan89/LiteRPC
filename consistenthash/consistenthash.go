package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

type ConsistentHash struct {
	replica int
	keys    []int // sorted
	hashMap map[int]string
}

func NewConsistentHash(replica int) *ConsistentHash {
	return &ConsistentHash{
		replica: replica,
		hashMap: make(map[int]string),
	}
}

// Add key to hash
func (c *ConsistentHash) Add(key string) {
	for i := 0; i < c.replica; i++ {
		hash := int(crc32.ChecksumIEEE([]byte(strconv.Itoa(i) + key)))
		c.keys = append(c.keys, hash)
		c.hashMap[hash] = key
	}
	sort.Ints(c.keys)
}

// Get key
func (c *ConsistentHash) Get(key string) string {
	if len(c.keys) == 0 {
		return ""
	}
	hash := int(crc32.ChecksumIEEE([]byte(key)))
	idx := sort.SearchInts(c.keys, hash)
	return c.hashMap[c.keys[idx%len(c.keys)]]
}

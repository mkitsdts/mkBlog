package bloom

import (
	"mkBlog/models"
	"mkBlog/pkg/database"
	"sync"
)

type BloomFilter struct {
	bitset   []byte
	bitcount []uint8
	hashs    []func(data []byte) int
}

const (
	// 256 bits = 32 bytes
	byteSize uint32 = 32
	bitSize  uint32 = byteSize * 8
)

var (
	bf   *BloomFilter
	once sync.Once
)

func Init() {
	bf = &BloomFilter{
		bitset:   make([]byte, byteSize),
		bitcount: make([]uint8, bitSize),
		hashs: []func(data []byte) int{
			func(data []byte) int {
				var h uint32 = 2166136261
				for _, b := range data {
					h = (h ^ uint32(b)) * 16777619 // FNV-1a
				}
				return int(h % bitSize)
			},
			func(data []byte) int {
				var h uint32 = 0
				for i, b := range data {
					h += uint32(b) * uint32(i+1)
				}
				return int(h % bitSize)
			},
			func(data []byte) int {
				var h uint32 = 0
				for i, b := range data {
					h += uint32(b) * uint32((i+1)*(i+1))
				}
				return int(h % bitSize)
			},
		},
	}
	var titles []string
	db := database.GetDatabase()
	if db == nil {
		println("bloom filter database error")
		return
	}
	println("bloom filter database ok")
	database.GetDatabase().Model(models.ArticleDetail{}).Select("title").Find(&titles)
	for _, title := range titles {
		bf.Add([]byte(title))
	}
}

func GetBloomFilter() *BloomFilter {
	once.Do(Init)
	return bf
}

func (bf *BloomFilter) Add(data []byte) {
	for _, h := range bf.hashs {
		pos := h(data)
		bytePos := pos / 8
		bitPos := pos % 8
		bf.bitset[bytePos] |= (1 << uint(bitPos))
		if bf.bitcount[pos] < 255 {
			bf.bitcount[pos]++
		}
	}
}

func (bf *BloomFilter) Exists(data []byte) bool {
	for _, h := range bf.hashs {
		pos := h(data)
		bytePos := pos / 8
		bitPos := pos % 8
		if (bf.bitset[bytePos] & (1 << uint(bitPos))) == 0 {
			return false
		}
	}
	return true
}

func (bf *BloomFilter) Remove(data []byte) {
	for _, h := range bf.hashs {
		pos := h(data)
		bytePos := pos / 8
		bitPos := pos % 8
		if bf.bitcount[pos] > 0 {
			bf.bitcount[pos]--
			if bf.bitcount[pos] == 0 {
				bf.bitset[bytePos] &^= 1 << uint(bitPos)
			}
		}
	}
}

func (bf *BloomFilter) Clear() {
	for i := range bf.bitset {
		bf.bitset[i] = 0
	}
	for i := range bf.bitcount {
		bf.bitcount[i] = 0
	}
}

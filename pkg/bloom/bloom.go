package bloom

import (
	"mkBlog/models"
	"mkBlog/pkg/database"
)

type BloomFilter struct {
	bitset   []byte
	bitcount []uint8
	hashs    []func(data []rune) uint8
}

const (
	// 256 bits = 32 bytes
	byteSize uint8 = 32
	bitSize  uint8 = 255
)

var bf *BloomFilter

func init() {
	bf = &BloomFilter{
		bitset:   make([]byte, byteSize),
		bitcount: make([]uint8, byteSize),
		hashs: []func(data []rune) uint8{
			func(data []rune) uint8 {
				var hash uint32 = 0
				for _, b := range data {
					hash += uint32(b)
				}
				return uint8(hash % uint32(bitSize))
			},
			func(data []rune) uint8 {
				var hash uint32 = 0
				for i, b := range data {
					hash += uint32(b) * uint32(i+1)
				}
				return uint8(hash % uint32(byteSize))
			},
			func(data []rune) uint8 {
				var hash uint32 = 0
				for i, b := range data {
					hash += uint32(b) * uint32((i+1)*(i+1))
				}
				return uint8(hash % uint32(byteSize))
			},
		},
	}
	var titles []string
	database.GetDatabase().Model(models.ArticleDetail{}).Select("title").Find(&titles)
	for _, title := range titles {
		bf.Add([]rune(title))
	}
}

func GetBloomFilter() *BloomFilter {
	return bf
}

func (bf *BloomFilter) Add(data []rune) {
	for _, h := range bf.hashs {
		pos := h(data)
		bytePos := pos / 8
		bitPos := pos % 8
		if bf.bitcount[bytePos] < 8 {
			bf.bitset[bytePos] |= (1 << bitPos)
			bf.bitcount[bytePos]++
		}
	}
}

func (bf *BloomFilter) Exists(data []rune) bool {
	for _, h := range bf.hashs {
		pos := h(data)
		bytePos := pos / 8
		bitPos := pos % 8
		if (bf.bitset[bytePos]&(1<<bitPos)) == 0 || bf.bitcount[bytePos] == 0 {
			return false
		}
	}
	return true
}

func (bf *BloomFilter) Remove(data []rune) {
	for _, h := range bf.hashs {
		pos := h(data)
		bytePos := pos / 8
		bitPos := pos % 8
		if bf.bitcount[bytePos] > 0 {
			bf.bitset[bytePos] &^= (1 << bitPos)
			bf.bitcount[bytePos]--
		}
	}
}

func (bf *BloomFilter) Clear() {
	for i := range bf.bitset {
		bf.bitset[i] = 0
		bf.bitcount[i] = 0
	}
}

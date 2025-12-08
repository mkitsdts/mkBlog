package bloom

import (
	"testing"
)

func TestBloomFilter(t *testing.T) {
	bloomfilter := GetBloomFilter()
	if bloomfilter == nil {
		println("bloom filter start error")
	}
	test_data := [][]byte{
		[]byte("test1"),
		[]byte("test2"),
		[]byte("test3"),
	}
	for _, data := range test_data {
		bloomfilter.Add(data)
		if !bloomfilter.Exists(data) {
			t.Errorf("Expected %s to be in the bloom filter", string(data))
		}
		if bloomfilter.Exists([]byte("not_in_filter")) {
			t.Error("Expected 'not_in_filter' to not be in the bloom filter")
		}
		if bloomfilter.Exists([]byte("1test")) {
			t.Error("Expected '1test' to not be in the bloom filter")
		}
		if bloomfilter.Exists([]byte("2test")) {
			t.Error("Expected '2test' to not be in the bloom filter")
		}
		if bloomfilter.Exists([]byte("3test")) {
			t.Error("Expected '3test' to not be in the bloom filter")
		}
	}
	for _, data := range test_data {
		bloomfilter.Remove(data)
		if bloomfilter.Exists(data) {
			t.Errorf("Expected %s to be removed from the bloom filter", string(data))
		}
	}
	if bloomfilter.Exists([]byte("test1")) || bloomfilter.Exists([]byte("test2")) || bloomfilter.Exists([]byte("test3")) {
		t.Error("Expected all test data to be removed from the bloom filter")
	}
	if bloomfilter.Exists([]byte("not_in_filter")) {
		t.Error("Expected 'not_in_filter' to not be in the bloom filter after removal")
	}
	if bloomfilter.Exists([]byte("1test")) {
		t.Error("Expected '1test' to not be in the bloom filter after removal")
	}
	if bloomfilter.Exists([]byte("2test")) {
		t.Error("Expected '2test' to not be in the bloom filter after removal")
	}
	if bloomfilter.Exists([]byte("3test")) {
		t.Error("Expected '3test' to not be in the bloom filter after removal")
	}
	println("Bloom filter test passed")
}

func TestBloomFilterChinese(t *testing.T) {
	bloomfilter := GetBloomFilter()
	test_data := [][]byte{
		[]byte("MySQL索引下推"),
		[]byte("限流算法"),
		[]byte("测试3"),
	}
	for _, data := range test_data {
		bloomfilter.Add(data)
		if !bloomfilter.Exists(data) {
			t.Errorf("Expected %s to be in the bloom filter", string(data))
		}
		if bloomfilter.Exists([]byte("不在过滤器中")) {
			t.Error("Expected '不在过滤器中' to not be in the bloom filter")
		}
	}
	for _, data := range test_data {
		bloomfilter.Remove(data)
		if bloomfilter.Exists(data) {
			t.Errorf("Expected %s to be removed from the bloom filter", string(data))
		}
	}
	if bloomfilter.Exists([]byte("MySQL索引下推")) || bloomfilter.Exists([]byte("限流算法")) || bloomfilter.Exists([]byte("测试3")) {
		t.Error("Expected all test data to be removed from the bloom filter")
	}
	if bloomfilter.Exists([]byte("不在过滤器中")) {
		t.Error("Expected '不在过滤器中' to not be in the bloom filter after removal")
	}
	println("Bloom filter Chinese test passed")
}

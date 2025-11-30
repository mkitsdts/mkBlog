package bloom

import "testing"

func TestBloomFilter(t *testing.T) {
	bf := GetBloomFilter()
	test_data := [][]rune{
		[]rune("test1"),
		[]rune("test2"),
		[]rune("test3"),
	}
	for _, data := range test_data {
		bf.Add(data)
		if !bf.Exists(data) {
			t.Errorf("Expected %s to be in the bloom filter", string(data))
		}
		if bf.Exists([]rune("not_in_filter")) {
			t.Error("Expected 'not_in_filter' to not be in the bloom filter")
		}
		if bf.Exists([]rune("1test")) {
			t.Error("Expected '1test' to not be in the bloom filter")
		}
		if bf.Exists([]rune("2test")) {
			t.Error("Expected '2test' to not be in the bloom filter")
		}
		if bf.Exists([]rune("3test")) {
			t.Error("Expected '3test' to not be in the bloom filter")
		}
	}
	for _, data := range test_data {
		bf.Remove(data)
		if bf.Exists(data) {
			t.Errorf("Expected %s to be removed from the bloom filter", string(data))
		}
	}
	if bf.Exists([]rune("test1")) || bf.Exists([]rune("test2")) || bf.Exists([]rune("test3")) {
		t.Error("Expected all test data to be removed from the bloom filter")
	}
	if bf.Exists([]rune("not_in_filter")) {
		t.Error("Expected 'not_in_filter' to not be in the bloom filter after removal")
	}
	if bf.Exists([]rune("1test")) {
		t.Error("Expected '1test' to not be in the bloom filter after removal")
	}
	if bf.Exists([]rune("2test")) {
		t.Error("Expected '2test' to not be in the bloom filter after removal")
	}
	if bf.Exists([]rune("3test")) {
		t.Error("Expected '3test' to not be in the bloom filter after removal")
	}
	t.Log("Bloom filter test passed")
}

func TestBloomFilterChinese(t *testing.T) {
	bf := GetBloomFilter()
	test_data := [][]rune{
		[]rune("测试1"),
		[]rune("测试2"),
		[]rune("测试3"),
	}
	for _, data := range test_data {
		bf.Add(data)
		if !bf.Exists(data) {
			t.Errorf("Expected %s to be in the bloom filter", string(data))
		}
		if bf.Exists([]rune("不在过滤器中")) {
			t.Error("Expected '不在过滤器中' to not be in the bloom filter")
		}
	}
	for _, data := range test_data {
		bf.Remove(data)
		if bf.Exists(data) {
			t.Errorf("Expected %s to be removed from the bloom filter", string(data))
		}
	}
	if bf.Exists([]rune("测试1")) || bf.Exists([]rune("测试2")) || bf.Exists([]rune("测试3")) {
		t.Error("Expected all test data to be removed from the bloom filter")
	}
	if bf.Exists([]rune("不在过滤器中")) {
		t.Error("Expected '不在过滤器中' to not be in the bloom filter after removal")
	}
	t.Log("Bloom filter Chinese test passed")
}

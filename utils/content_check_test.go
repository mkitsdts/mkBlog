package utils

import "testing"

func TestContentCheck(t *testing.T) {
	tests := []string{
		"Hello, world!",
		"你好，世界！",
		"今日は世界",
		"Bonjour le monde!",
		"1234567890",
		"!@#$%^&*()_+",
	}
	expected := []bool{false, true, true, false, false, false}
	for i, test := range tests {
		result := ContainsCJK(test)
		if result != expected[i] {
			t.Errorf("ContainsCJK(%q) = %v; want %v", test, result, expected[i])
		}
	}
}

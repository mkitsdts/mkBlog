package utils

// 粗略判断查询字符串中是否包含中日韩统一表意文字（CJK Unified Ideographs）
func ContainsCJK(s string) bool {
	for _, r := range s {
		// 中英文范围：
		if (r >= 0x4E00 && r <= 0x9FFF) || // CJK Unified Ideographs
			(r >= 0x3400 && r <= 0x4DBF) || // CJK Unified Ideographs Extension A
			(r >= 0x20000 && r <= 0x2A6DF) || // Extension B
			(r >= 0x2A700 && r <= 0x2B73F) || // Extension C
			(r >= 0x2B740 && r <= 0x2B81F) || // Extension D
			(r >= 0x2B820 && r <= 0x2CEAF) || // Extension E
			(r >= 0xF900 && r <= 0xFAFF) ||
			(r >= 0x61 && r <= 0x7A) ||
			(r >= 0x41 && r <= 0x5A) { // CJK Compatibility Ideographs
			return true
		}
	}
	return false
}

package utils

func IsLegalComment(content string) bool {
	// 判断长度是否合法
	if len(content) == 0 || len(content) > 500 {
		return false
	}
	// 后续添加更多验证规则
	return true
}

package utils

import "strings"

func StripMarkdown(md string) string {
	// 简单去除一些常见的 Markdown 语法
	replacements := []struct {
		old string
		new string
	}{
		{"**", ""},     // 粗体
		{"*", ""},      // 斜体
		{"__", ""},     // 粗体
		{"_", ""},      // 斜体
		{"`", ""},      // 行内代码
		{"```", ""},    // 代码块
		{"#", ""},      // 标题
		{"##", ""},     // 二级标题
		{"###", ""},    // 三级标题
		{"####", ""},   // 四级标题
		{"#####", ""},  // 五级标题
		{"######", ""}, // 六级标题
		{"-", ""},      // 列表项
		{">", ""},      // 引用
		{"![", ""},     // 图片开始
		{"](", ""},     // 图片结束
		{"[", ""},      // 链接开始
		{")", ""},      // 链接结束
		{"\n", " "},    // 换行替换为空格
	}

	for _, r := range replacements {
		md = strings.ReplaceAll(md, r.old, r.new)
	}

	return md
}

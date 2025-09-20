package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"time"

	"gopkg.in/yaml.v3"
)

type Meta struct {
	Title    string `json:"title"`
	Category string `json:"category"`
	Summary  string `json:"summary"`
	Author   string `json:"author"`
	UpdateAt string `json:"update_at"`
	CreateAt string `json:"create_at"`
}

type Payload struct {
	Title    string `json:"title"`
	Category string `json:"category"`
	Summary  string `json:"summary"`
	Content  string `json:"content"`
	Author   string `json:"author"`
}

var (
	frontMatterRE = regexp.MustCompile(`(?s)^---\n(.*?)\n---\n(.*)$`)
	lineKVRE      = regexp.MustCompile(`^([A-Za-z_][A-Za-z0-9_-]*)\s*:\s*(.*)$`)
)

var blog_template string = `---
title: {{.Title}}
category: 
author: 
create_at: {{.CreateAt}}
---
`

// 预编译模板（程序启动时就检查语法）
var tmpl = template.Must(template.New("blog").Parse(blog_template))

// 生成文本
func generateContent(title string) (string, error) {
	now := time.Now().Format("2006-01-02 15:04:05")
	data := map[string]any{
		"Title":    title,
		"CreateAt": now,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func parseFile(path string) (Meta, string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Meta{}, "", err
	}
	text := string(b)
	m := frontMatterRE.FindStringSubmatch(text)
	if len(m) != 3 {
		return Meta{}, "", errors.New("缺少 front matter 块")
	}
	metaBlock := m[1]
	body := m[2]
	meta := Meta{}
	for _, line := range strings.Split(metaBlock, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		kv := lineKVRE.FindStringSubmatch(line)
		if len(kv) != 3 {
			continue
		}
		k := strings.ToLower(kv[1])
		v := strings.TrimSpace(kv[2])
		switch k {
		case "title":
			meta.Title = v
		case "category":
			meta.Category = v
		case "summary":
			meta.Summary = v
		case "author":
			meta.Author = v
		case "date":
			meta.CreateAt = v
		case "create_at":
			meta.CreateAt = v
		case "update_at":
			meta.UpdateAt = v
		}
	}
	if meta.Title == "" {
		return Meta{}, "", errors.New("title 必填")
	}
	if meta.Category == "" {
		meta.Category = "General"
	}
	if meta.Summary == "" {
		// 自动截取前 100 字符
		pure := strings.ReplaceAll(stripMarkdown(body), "\n", " ")
		if len(pure) > 100 {
			pure = pure[:100] + "..."
		}
		meta.Summary = pure
	}
	if meta.Author == "" {
		meta.Author = "null"
	}
	if meta.CreateAt == "" {
		meta.CreateAt = time.Now().Format("2006-01-02 15:04:05")
	}
	return meta, body, nil
}

func stripMarkdown(s string) string {
	// 简单去掉 markdown 语法
	s = regexp.MustCompile("`{1,3}.*?`{1,3}").ReplaceAllString(s, "")
	s = regexp.MustCompile(`!\[[^\]]*\]\([^)]+\)`).ReplaceAllString(s, "")
	s = regexp.MustCompile(`\[[^\]]*\]\([^)]+\)`).ReplaceAllString(s, "")
	s = regexp.MustCompile(`[#>*_\-~]`).ReplaceAllString(s, "")
	return strings.TrimSpace(s)
}

func pushOne(server, secret string, meta Meta, body string) error {
	payload := Payload{
		Title:    meta.Title,
		Category: meta.Category,
		Summary:  meta.Summary,
		Content:  body,
		Author:   meta.Author,
	}

	if server == "" {
		server = Cfg.Server
	}
	if secret == "" {
		secret = Cfg.Secret
	}

	// 假设后端接口：PUT /api/article/:title  JSON body
	url := fmt.Sprintf("%s/api/article/%s", strings.TrimRight(server, "/"), urlEscape(meta.Title))
	data, _ := json.Marshal(payload)
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if secret != "" {
		req.Header.Set("Authorization", "Bearer "+secret)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	rb, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return fmt.Errorf("上传失败 %s: %s", resp.Status, string(rb))
	}
	fmt.Printf("✅ %s 上传成功\n", meta.Title)
	return nil
}

// Image 定义与后端 models.Image 对齐
type Image struct {
	Title string `json:"title"`
	Data  []byte `json:"data"`
	Name  string `json:"name"`
}

var imageExts = map[string]struct{}{
	".png":  {},
	".jpg":  {},
	".jpeg": {},
	".webp": {},
	".svg":  {},
	".ico":  {},
}

// pushImages 扫描 markdown 同目录下同名文件夹中的图片并上传
func pushImages(server, secret, articleTitle, mdPath string) (int, error) {
	if server == "" {
		server = Cfg.Server
	}
	if secret == "" {
		secret = Cfg.Secret
	}
	dir := filepath.Dir(mdPath)
	base := filepath.Base(mdPath)
	dot := strings.LastIndex(base, ".")
	if dot > 0 {
		base = base[:dot]
	}
	imgDir := filepath.Join(dir, base)
	stat, err := os.Stat(imgDir)
	if err != nil || !stat.IsDir() {
		return 0, nil // 没有同名目录直接跳过
	}
	entries, err := os.ReadDir(imgDir)
	if err != nil {
		return 0, err
	}
	count := 0
	for _, e := range entries {
		if e.IsDir() { // 不递归
			continue
		}
		name := e.Name()
		ext := strings.ToLower(filepath.Ext(name))
		if _, ok := imageExts[ext]; !ok {
			continue
		}
		full := filepath.Join(imgDir, name)
		data, err := os.ReadFile(full)
		if err != nil {
			fmt.Printf("  ⚠️  读取图片失败 %s: %v\n", name, err)
			continue
		}
		img := &Image{Title: articleTitle, Name: name, Data: data}
		url := fmt.Sprintf("%s/api/image", strings.TrimRight(server, "/"))
		jb, _ := json.Marshal(img)
		req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(jb))
		if err != nil {
			fmt.Printf("  ❌ 构建请求失败 %s: %v\n", name, err)
			continue
		}
		req.Header.Set("Content-Type", "application/json")
		if secret != "" {
			req.Header.Set("Authorization", "Bearer "+secret)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Printf("  ❌ 上传失败 %s: %v\n", name, err)
			continue
		}
		rb, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode >= 300 {
			fmt.Printf("  ❌ 上传失败 %s: %s %s\n", name, resp.Status, string(rb))
			continue
		}
		fmt.Printf("  🖼️  %s 上传成功\n", name)
		count++
	}
	if count > 0 {
		fmt.Printf("  共上传图片 %d 张 (文章: %s)\n", count, articleTitle)
	}
	return count, nil
}

func deleteArticle(title string) error {
	url := fmt.Sprintf("%s/api/article/%s", strings.TrimRight(Cfg.Server, "/"), urlEscape(title))
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if Cfg.Secret != "" {
		req.Header.Set("Authorization", "Bearer "+Cfg.Secret)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	rb, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return fmt.Errorf("删除失败 %s: %s", resp.Status, string(rb))
	}
	fmt.Printf("✅ %s 删除成功\n", title)
	return nil
}

func urlEscape(s string) string {
	repl := strings.ReplaceAll(s, " ", "%20")
	return repl
}

// 处理 push 子命令
func runPush(args []string) error {
	fs := flag.NewFlagSet("push", flag.ContinueOnError)
	server := fs.String("server", "", "后端服务地址(可用 MKBLOG_SERVER)")
	token := fs.String("token", "", "鉴权 token(可用 MKBLOG_TOKEN)")
	ext := fs.String("ext", ".md", "文章文件扩展名")
	skipHidden := fs.Bool("skip-hidden", true, "跳过隐藏文件/目录")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	count := 0
	err = filepath.Walk(cwd, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		name := info.Name()
		if *skipHidden && strings.HasPrefix(name, ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(strings.ToLower(name), strings.ToLower(*ext)) {
			return nil
		}
		meta, body, perr := parseFile(p)
		if perr != nil {
			fmt.Printf("跳过 %s: %v\n", p, perr)
			return nil
		}
		meta.UpdateAt = info.ModTime().Format("2006-01-02 15:04:05")
		if err := pushOne(*server, *token, meta, body); err != nil {
			fmt.Printf("❌ %s 失败: %v\n", meta.Title, err)
		} else {
			// 上传成功后尝试上传图片
			_, _ = pushImages(*server, *token, meta.Title, p)
			count++
		}
		return nil
	})
	if err != nil {
		return err
	}
	if count == 0 {
		fmt.Println("未找到文章文件")
	} else {
		fmt.Printf("完成：%d 篇已处理\n", count)
	}
	return nil
}

func runCreate(args []string) error {
	fs := flag.NewFlagSet("create", flag.ContinueOnError)
	title := fs.String("title", "", "文章标题")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *title == "" {
		return fmt.Errorf("文章标题不能为空")
	}

	content, err := generateContent(*title)
	if err != nil {
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	filename := *title + ".md"
	err = os.WriteFile(filepath.Join(cwd, filename), []byte(content), 0644)
	if err != nil {
		return err
	}

	fmt.Printf("创建文章: %s\n", *title)
	return nil
}

func runDelete(args []string) error {
	fs := flag.NewFlagSet("delete", flag.ContinueOnError)
	title := fs.String("title", "", "文章标题")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *title == "" {
		return fmt.Errorf("文章标题不能为空")
	}

	// 发送删除请求
	if err := deleteArticle(*title); err != nil {
		return err
	}

	fmt.Printf("删除文章: %s\n", *title)
	return nil
}

func usage() {
	fmt.Println(`mkblog 用法:
  mkblog push [选项]      扫描当前目录及子目录上传 Markdown

通用选项( push ):
  -server URL            后端地址 (默认环境 MKBLOG_SERVER 或 http://localhost:8080)
  -token TOKEN           鉴权 Token (或环境 MKBLOG_TOKEN)
  -ext .md               文章扩展名
  -skip-hidden true      跳过以 . 开头的文件和目录

示例:
  mkblog push
  mkblog push -server http://blog.example.com -token ABC123
  MKBLOG_SERVER=http://remote:8080 mkblog push`)
}

type Config struct {
	Server string `json:"server" yaml:"server"`
	Secret string `json:"secret" yaml:"secret"`
}

var Cfg *Config = &Config{}

func LoadConfig() error {
	data, err := os.ReadFile("configs.yaml")
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal(data, Cfg); err != nil {
		return err
	}
	return nil
}

func main() {

	if err := LoadConfig(); err != nil {
		fmt.Println("加载配置失败:", err)
		os.Exit(1)
	}

	fmt.Println("mkBlog CLI 工具", Cfg.Server)

	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}
	sub := os.Args[1]
	switch sub {
	case "push":
		if err := runPush(os.Args[2:]); err != nil {
			fmt.Println("错误:", err)
			os.Exit(1)
		}
	case "create":
		if err := runCreate(os.Args[2:]); err != nil {
			fmt.Println("错误:", err)
			os.Exit(1)
		}
	case "delete":
		if err := runDelete(os.Args[2:]); err != nil {
			fmt.Println("错误:", err)
			os.Exit(1)
		}
	default:
		usage()
		os.Exit(1)
	}
}

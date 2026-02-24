package config

import (
	"log/slog"
	"mkBlog/models"
	"os"
	"path"
)

const impl string = `# 所有路径都是与程序同一目录下的data文件夹下的相对路径
database:
  host: app.db   # 数据库地址，如果启用sqlite3，只需要在 host 填写路径，其他不用管
  port: 3306          # 数据库端口
  user: root          # 数据库用户
  password: root      # 数据库密码
  name: mkblog        # 数据库名称
  kind: sqlite3       # 数据库类型，支持mysql和postgres以及sqlite3

tls:
  enabled: false              # 是否启用TLS
  cert: static/server.crt   # TLS证书文件路径
  key: static/server.key    # TLS密钥文件路径

cert_control:
  email: 114514@colima.com      # 注册 Let's Encrypt 账号的邮箱
  domain: 114514.com            # 需要申请 TLS 证书的域名
  Key:                          # 域名提供商的 API Key
  Secret:                       # 域名提供商的 API 密钥
  DomainProvider: Aliyun        # 域名提供商。可选： Aliyun , TencentCloud

auth:
  enabled: false                        # 是否启用身份验证
  secret: 123456789015234564892545456 # 自行设置验证密码

server:
  port: 4801                     # 服务器端口
  imageSavePath: static/images   # 图片保存路径
  limiter:
    requests: 100                # 每个IP在duration内最多允许的请求数
    duration: 5                  # 限制的时间窗口，单位为秒
  devmode: true
  http3_enabled: false
  cert_ctrl_enabled: false

site:
  signature: "鼠鼠是穿越者..."         # 个性签名
  about: "鼠鼠是一个喜欢折腾的程序员..." # 关于我
  avatarPath: avatar.jpg          # 头像路径
  comment_enabled: false          # 是否启用评论功能
  icp: 粤ICP备123456789号          # 备案号（可选）
`

func PWD() string {
	d, err := os.Getwd()
	if err != nil {
		d = "."
	}
	return d
}

func writeImpl() error {
	slog.Debug("write impl begin")
	p := path.Join(models.Default_Data_Path, models.Default_Config_File_Path)
	return os.WriteFile(p, []byte(impl), 0644)
}

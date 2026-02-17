package config

import (
	"mkBlog/models"
	"os"
)

const impl string = `# 所有路径都是与程序同一目录下的data文件夹下的相对路径
database:
  host: data/app.db   # 数据库地址，如果启用sqlite3，只需要在 host 填写路径，其他不用管
  port:           	  # 数据库端口
  user:           	  # 数据库用户
  password:           # 数据库密码
  name: mkblog        # 数据库名称
  kind: sqlite3       # 数据库类型，支持mysql和postgres以及sqlite3

tls:
  enabled: false              # 是否启用TLS
  cert: ./static/server.crt   # TLS证书文件路径
  key: ./static/server.key    # TLS密钥文件路径

cert_control:
  enabled: false				# 是否启用自动 TLS 证书管理
  email: 114514@colima.com      # 注册 Let's Encrypt 账号的邮箱
  domain: 114514.com            # 需要申请 TLS 证书的域名
  Key:                          # 域名提供商的 API Key
  Secret:                       # 域名提供商的 API 密钥
  DomainProvider: Aliyun        # 域名提供商。可选： Aliyun , TencentCloud

auth:
  enabled: false                        # 是否启用身份验证
  secret: "123456789015234564892545456" # 自行设置验证密码

server:
  port: 4801                     # 服务器端口
  imageSavePath: ./static/images # 图片保存路径
  limiter:
    requests: 100                # 每个IP在duration内最多允许的请求数
    duration: 5                  # 限制的时间窗口，单位为秒
  devmode: false
  http3_enabled: false

site:
  signature: "emm......."        # 个性签名
  about: "鼠鼠我................" # 关于我
  avatarPath: avatar.jpg         # 头像路径
  server: http://localhost:4801  # 服务器地址(端口为80/443时后面可以不加端口号)
  devmode: false                 # 开发模式，启用后前端使用未压缩的资源文件
  comment_enabled: false         # 是否启用评论功能
  icp: 粤ICP备123456789号         # 备案号（可选）
`

func writeImpl() error {
	return os.WriteFile(models.Default_Config_File_Path, []byte(impl), 0644)
}

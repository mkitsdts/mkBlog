package sslcontroler

import (
	"context"
	"log/slog"
	"mkBlog/config"
	"time"
)

const original_cert_id_path = "tls_cert_id"

type SSLControler struct {
	domain         string // 绑定的域名
	domain_manager SSLService
	cert_manager   SSLService
	current_status int
}

var current_cert_id string
var sslControler *SSLControler

func init() {
	if !config.Cfg.TLS.Enabled {
		return
	}
	sslControler = &SSLControler{}

	sslControler.domain = config.Cfg.Site.Server
	// 初始化证书和域名提供者
	switch config.Cfg.TLS.CertProvider {
	case "tencentcloud":
		var err error
		sslControler.cert_manager, err = NewTencentCloudService(
			&TencentCloudServiceOption{
				SecretId:  config.Cfg.TLS.CertProviderKey,
				SecretKey: config.Cfg.TLS.CertProviderSecret,
				Domain:    config.Cfg.TLS.Domain,
			},
		)
		if err != nil {
			panic("tencentcloud SSLService init failed: " + err.Error())
		}
	case "alicloud":
		var err error
		sslControler.cert_manager, err = NewALiCloudService(
			&ALiCloudServiceOption{
				AccessKeyId:     config.Cfg.TLS.CertProviderKey,
				AccessKeySecret: config.Cfg.TLS.CertProviderSecret,
				Domain:          config.Cfg.TLS.Domain,
			},
		)
		if err != nil {
			panic("alicloud SSLService init failed: " + err.Error())
		}
	}
	switch config.Cfg.TLS.DomainProvider {
	case "tencentcloud":
		var err error
		sslControler.domain_manager, err = NewTencentCloudService(
			&TencentCloudServiceOption{
				SecretId:  config.Cfg.TLS.DomainProviderKey,
				SecretKey: config.Cfg.TLS.DomainProviderSecret,
				Domain:    config.Cfg.TLS.Domain,
			},
		)
		if err != nil {
			panic("tencentcloud SSLService init failed: " + err.Error())
		}
	case "alicloud":
		var err error
		sslControler.domain_manager, err = NewALiCloudService(
			&ALiCloudServiceOption{
				AccessKeyId:     config.Cfg.TLS.DomainProviderKey,
				AccessKeySecret: config.Cfg.TLS.DomainProviderSecret,
				Domain:          config.Cfg.TLS.Domain,
			},
		)
		if err != nil {
			panic("alicloud SSLService init failed: " + err.Error())
		}
	}
	if config.Cfg.TLS.OriginalCertID == "" {
		// 如果没有配置原始证书ID，则不启动控制器
		slog.Warn("SSLControler: TLS is enabled but original_cert_id is not set, SSLControler will not start")
		return
	}
	current_cert_id = config.Cfg.TLS.OriginalCertID

	content := []byte(config.Cfg.TLS.OriginalCertID)
	SaveFile(original_cert_id_path, &content)
}

func (c *SSLControler) Start(ctx context.Context) {
	duration := time.Duration(config.Cfg.TLS.CheckInterval) * time.Hour
	ticker := time.NewTicker(duration)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			slog.Info("SSLControler: Starting SSL certificate check")
			c.checkCurrStatus()
		case <-ctx.Done():
			return
		default:
			c.handleState()
			time.Sleep(10 * time.Second)
		}
	}
}

func (c *SSLControler) checkCurrStatus() {
	// 检查当前证书状态
	info, err := c.domain_manager.CheckDVAuthStatus(config.Cfg.TLS.OriginalCertID)
	if err != nil {
		slog.Error("SSLControler: Failed to check certificate status", "error", err)
		c.current_status = FAILED_TO_GET_STATUS
		return
	}
	if info == nil {
		c.current_status = STATUS_UNKNOWN
		return
	}
	var certEnd time.Time
	// 先尝试 RFC3339
	if s, ok := any(info.CertEndTime).(string); ok {
		certEnd, err = time.Parse(time.RFC3339, s)
		if err != nil {
			// 如果是 "2006-01-02 15:04:05" 这种格式，再尝试解析（按需调整布局）
			certEnd, err = time.ParseInLocation("2006-01-02 15:04:05", s, time.Local)
		}
		if err != nil {
			slog.Error("SSLControler: Failed to parse CertEndTime", "value", s, "error", err)
		}
	} else if t, ok := any(info.CertEndTime).(time.Time); ok {
		certEnd = t
	}

	if certEnd.IsZero() {
		c.current_status = FAILED_TO_GET_STATUS
		return
	}

	// 用 time.Until 更直观：距离到期小于等于 14 天则续证
	if time.Until(certEnd) <= 14*24*time.Hour {
		c.current_status = NEED_RENEWAL
	} else {
		c.current_status = STATUS_VALID
	}
}

func RunSSLControler(ctx context.Context) {
	if sslControler == nil {
		return
	}
	go sslControler.Start(ctx)
}

func (c *SSLControler) applyCertificate() error {
	c.current_status = APPLYING_CERTIFICATE
	// 申请新证书
	certID, err := c.cert_manager.ApplyCertificate()
	if err != nil {
		slog.Error("SSLControler: Failed to apply for new certificate", "error", err)
		c.current_status = APPLYING_CERTIFICATE_FAILED
		return err
	}
	current_cert_id = certID
	c.current_status = APPLYING_CERTIFICATE
	certid_content := []byte(current_cert_id)
	SaveFile(original_cert_id_path, &certid_content)
	return nil
}

// 完成证书验证
func (c *SSLControler) validCertificate() {
	// 获取DV认证详情
	dns, err := c.cert_manager.GetDVAuthDetail(current_cert_id)
	if err != nil {
		slog.Error("SSLControler: Failed to get DV auth detail", "error", err)
		return
	}
	time.Sleep(1 * time.Minute) // 等待一会儿，确保 DNS 记录生效

	c.current_status = ADDING_DOMAIN_RECORD
	// 添加域名解析记录
	resp, err := c.domain_manager.AddDomainRecord("_dnsauth", "TXT", dns)
	if err != nil {
		slog.Error("SSLControler: Failed to add domain record", "error", err)
		c.current_status = ADD_DOMAIN_RECORD_FAILED
		return
	}
	if resp == "" {
		slog.Error("SSLControler: Empty response when adding domain record")
		c.current_status = ADD_DOMAIN_RECORD_FAILED
		return
	}
	c.current_status = VALID_DOMAIN_SUCCESS
}

func (c *SSLControler) downloadCertificate() error {
	// 下载证书
	cert, key, err := c.cert_manager.GetCertificate(current_cert_id)
	if err != nil {
		slog.Error("SSLControler: Failed to get certificate", "error", err)
		return err
	}

	// 保存证书到本地
	SaveFile(config.Cfg.TLS.Cert, cert.Data)
	SaveFile(config.Cfg.TLS.Key, key.Data)

	return nil
}

func (c *SSLControler) RenewCertificate() {
	if err := c.applyCertificate(); err != nil {
		slog.Error("SSLControler: Failed to renew certificate", "error", err)
		c.current_status = APPLYING_CERTIFICATE
		return
	}
	time.Sleep(1 * time.Minute) // 等待一会儿，确保申请生效
	c.validCertificate()
}

func (c *SSLControler) handleState() {
	switch c.current_status {
	case NEED_RENEWAL:
		slog.Info("SSLControler: Certificate needs renewal, starting renewal process")
		c.RenewCertificate()
	case APPLYING_CERTIFICATE, APPLYING_CERTIFICATE_FAILED:
		slog.Info("SSLControler: Certificate is being issued, waiting")
		c.applyCertificate()
	case ADDING_DOMAIN_RECORD, ADD_DOMAIN_RECORD_FAILED, VALID_DOMAIN_FAILED, VALIDATING_DOMAIN:
		slog.Info("SSLControler: Adding domain record, waiting")
		c.validCertificate()
	case VALID_DOMAIN_SUCCESS:
		slog.Info("SSLControler: Domain validated successfully, downloading certificate")
		if err := c.downloadCertificate(); err != nil {
			slog.Error("SSLControler: Failed to download certificate", "error", err)
			return
		}
		slog.Info("SSLControler: Certificate downloaded and saved successfully")
		c.current_status = STATUS_VALID
	default:
		// 其他状态不处理
	}
}

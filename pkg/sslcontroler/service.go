package sslcontroler

// 申请免费 SSL 证书服务的统一接口
type SSLService interface {
	// 申请 SSL 证书，返回证书 ID
	ApplyCertificate() (string, error)
	// 获取 SSL 证书需要添加的 DNS 记录值，用于完成域名所有权验证
	GetDVAuthDetail(id string) (string, error)
	// 检查域名所有权验证状态
	CheckDVAuthStatus(id string) (*CertInfo, error)
	// 下载 SSL 证书，返回公钥和私钥
	GetCertificate(id string) (*Certificate, *CertificateKey, error)
	// 添加 DNS 解析记录
	AddDomainRecord(rr, tp, dv string) (string, error)
}

func ApplyCertificate(service SSLService) (string, error) {
	return service.ApplyCertificate()
}

func GetDVAuthDetail(service SSLService, id string) (string, error) {
	return service.GetDVAuthDetail(id)
}

func CheckDVAuthStatus(service SSLService, id string) (*CertInfo, error) {
	return service.CheckDVAuthStatus(id)
}

func GetCertificate(service SSLService, id string) (*Certificate, *CertificateKey, error) {
	return service.GetCertificate(id)
}

package tlscert

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"time"
)

// renewBefore 是提前续期的时间窗口，证书剩余有效期小于它就开始续期。
const renewBefore = 14 * 24 * time.Hour

// certState 是磁盘上证书文件的状态。
//
// 只有“证书肯定不存在”才算 CertMissing。读不到、解析不了、证书与私钥
// 不匹配都属于异常状态，此时绝不自动签发新证书：那会白白消耗
// Let's Encrypt 的重复证书配额，也不能真正修好文件。
type certState int

const (
	// CertValid 证书可用且不在续期窗口内。
	CertValid certState = iota

	// CertExpiring 证书可用，但剩余有效期不足 renewBefore。
	CertExpiring

	// CertMissing 证书文件不存在，通常是首次部署，需要签发。
	CertMissing
)

func (s certState) String() string {
	switch s {
	case CertValid:
		return "valid"
	case CertExpiring:
		return "expiring"
	case CertMissing:
		return "missing"
	default:
		return fmt.Sprintf("certState(%d)", int(s))
	}
}

// certInfo 是证书检查结果。
type certInfo struct {
	State    certState
	NotAfter time.Time
	Domains  []string
}

// inspectTLSCert 检查证书与私钥文件的状态。
func inspectTLSCert(certPath, keyPath string, now time.Time) (certInfo, error) {
	certPEM, err := os.ReadFile(certPath)
	if errors.Is(err, fs.ErrNotExist) {
		return certInfo{State: CertMissing}, nil
	}
	if err != nil {
		return certInfo{}, fmt.Errorf("read certificate %s: %w", certPath, err)
	}

	keyPEM, err := os.ReadFile(keyPath)
	if err != nil {
		return certInfo{}, fmt.Errorf("read private key %s: %w", keyPath, err)
	}

	leaf, err := firstCertificate(certPEM)
	if err != nil {
		return certInfo{}, fmt.Errorf("parse certificate %s: %w", certPath, err)
	}

	// 证书和私钥必须成对可用，否则服务起不来；这种情况要人工介入，
	// 自动重新签发只会在错误的路径上反复烧配额。
	if _, err := tls.X509KeyPair(certPEM, keyPEM); err != nil {
		return certInfo{}, fmt.Errorf("certificate %s does not match private key %s, delete both files to get a new pair: %w", certPath, keyPath, err)
	}

	info := certInfo{State: CertValid, NotAfter: leaf.NotAfter, Domains: leaf.DNSNames}
	if now.Add(renewBefore).After(leaf.NotAfter) {
		info.State = CertExpiring
	}
	return info, nil
}

func firstCertificate(certPEM []byte) (*x509.Certificate, error) {
	rest := certPEM
	for {
		var block *pem.Block
		block, rest = pem.Decode(rest)
		if block == nil {
			return nil, errors.New("no CERTIFICATE block found")
		}
		if block.Type != "CERTIFICATE" {
			continue
		}
		return x509.ParseCertificate(block.Bytes)
	}
}

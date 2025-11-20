package sslcontroler

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	ssl "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/ssl/v20191205"
)

type TencentCloudService struct {
	client       *ssl.Client
	applyrequest *ssl.ApplyCertificateRequest
	domain       string
}

type TencentCloudServiceOption struct {
	SecretId  string
	SecretKey string
	Region    string
	Domain    string
}

func NewTencentCloudService(option *TencentCloudServiceOption) (*TencentCloudService, error) {
	if option.SecretId == "" || option.SecretKey == "" || option.Domain == "" {
		return nil, fmt.Errorf("TencentCloudService Init failed: parameter missing")
	}

	t := &TencentCloudService{}
	credential := common.NewCredential(
		option.SecretId,
		option.SecretKey,
	)

	if option.Region == "" {
		option.Region = Guangzhou
	}
	t.domain = option.Domain

	var err error
	t.client, err = ssl.NewClient(credential, option.Region, profile.NewClientProfile())
	if err != nil {
		return nil, err
	}

	t.applyrequest = ssl.NewApplyCertificateRequest()
	t.applyrequest.DomainName = &t.domain
	t.applyrequest.DvAuthMethod = common.StringPtr("DNS_AUTO")

	return t, nil
}

func (t *TencentCloudService) ApplyCertificate() (string, error) {
	resp, err := t.client.ApplyCertificate(t.applyrequest)
	if err != nil {
		return "", err
	}
	return *resp.Response.CertificateId, nil
}

func (t *TencentCloudService) GetCertificate(id string) (*Certificate, *CertificateKey, error) {
	request := ssl.NewDownloadCertificateRequest()
	request.CertificateId = &id

	resp, err := t.client.DownloadCertificate(request)
	if err != nil {
		return nil, nil, err
	}
	if resp.Response.Content == nil {
		return nil, nil, fmt.Errorf("get certificate failed")
	}

	cert := &Certificate{}
	key := &CertificateKey{}

	// content is zip file in base64 encoding
	zipData, err := base64.StdEncoding.DecodeString(*resp.Response.Content)
	if err != nil {
		return nil, nil, fmt.Errorf("decode base64 zip content failed: %w", err)
	}

	zr, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return nil, nil, fmt.Errorf("open zip reader failed: %w", err)
	}

	// check each file in zip
	files := map[string][]byte{}
	for _, f := range zr.File {
		rc, err := f.Open()
		if err != nil {
			return nil, nil, fmt.Errorf("open zip entry %s failed: %w", f.Name, err)
		}
		b, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return nil, nil, fmt.Errorf("read zip entry %s failed: %w", f.Name, err)
		}
		name := filepath.Base(f.Name)
		files[name] = b
	}

	// extract cert, key, ca from files
	var certPEM, keyPEM, caPEM []byte
	for name, b := range files {
		l := strings.ToLower(name)
		switch {
		case strings.HasSuffix(l, ".key") || strings.Contains(l, "private"):
			keyPEM = b
		case strings.HasSuffix(l, ".pem") || strings.HasSuffix(l, ".crt") || strings.HasSuffix(l, ".cer"):
			// distinguish cert and ca by filename
			if strings.Contains(l, "ca") || strings.Contains(l, "bundle") || strings.Contains(l, "chain") {
				caPEM = b
			} else if certPEM == nil {
				certPEM = b
			} else {
				// multiple cert files, choose the larger one as cert, smaller one as ca
				if len(b) > len(certPEM) {
					caPEM = certPEM
					certPEM = b
				} else {
					caPEM = b
				}
			}
		default:
			// other files
			slog.Info("Ignoring unrecognized file in zip:", slog.String("filename", name))
		}
	}

	if certPEM == nil || keyPEM == nil {
		return nil, nil, fmt.Errorf("certificate or private key not found in zip")
	}

	// 合并 cert + ca 为完整证书链（如果存在 ca）
	fullCert := certPEM
	if caPEM != nil {
		fullCert = append(fullCert, []byte("\n")...)
		fullCert = append(fullCert, caPEM...)
	}

	// 将内容填入 autossl 类型（根据你的 autossl 包的字段名做相应调整）
	cert.Data = &fullCert
	cert.Name = t.domain + ".crt"
	key.Data = &keyPEM
	key.Name = t.domain + ".key"

	return cert, key, nil
}

const (
	CERT_STATUS_ISSUED     = 0
	CERT_STATUS_IN_ISSUING = 1
)

func (t *TencentCloudService) CheckCertificate(id string) (bool, error) {
	request := ssl.NewDescribeCertificatesRequest()
	request.CertIds = []*string{&id}

	resp, err := t.client.DescribeCertificates(request)
	if err != nil {
		return false, err
	}
	if len(resp.Response.Certificates) == 0 {
		return false, fmt.Errorf("certificate not found")
	}
	fmt.Println("Certificate status:", *resp.Response.Certificates[0].Status)
	if *resp.Response.Certificates[0].CertificateId == id {
		// check status
		switch *resp.Response.Certificates[0].Status {
		case CERT_STATUS_ISSUED:
			return true, nil
		case CERT_STATUS_IN_ISSUING:
			slog.Info("Certificate still being issued", slog.String("status", fmt.Sprintf("%d", *resp.Response.Certificates[0].Status)))
			return false, fmt.Errorf("being issued")
		default:
			slog.Info("Certificate not issued yet", slog.String("status", fmt.Sprintf("%d", *resp.Response.Certificates[0].Status)))
			return false, nil
		}
	}

	return false, fmt.Errorf("certificate ID mismatch")
}

func (t *TencentCloudService) AddDomainRecord(rr, tp, dv string) (string, error) {
	// TencentCloud does not support adding domain records via SSL API
	return "", fmt.Errorf("TencentCloudService AddDomainRecord not implemented")
}

func (t *TencentCloudService) GetDVAuthDetail(id string) (string, error) {
	request := ssl.NewDescribeCertificateDetailRequest()
	request.CertificateId = common.StringPtr(id)

	resp, err := t.client.DescribeCertificateDetail(request)
	if err != nil {
		return "", err
	}
	return *resp.Response.DvAuthDetail.DvAuthValue, nil
}

func (t *TencentCloudService) CheckDVAuthStatus(id string) (*CertInfo, error) {
	request := ssl.NewDescribeCertificateDetailRequest()
	request.CertificateId = common.StringPtr(id)

	resp, err := t.client.DescribeCertificateDetail(request)
	if err != nil {
		return nil, err
	}
	if *resp.Response.Status == CERT_STATUS_ISSUED {
		return &CertInfo{
			Status:      1,
			CertEndTime: *resp.Response.CertEndTime,
			DvDetail: &DvAuthDetail{
				DvAuthKey:   *resp.Response.DvAuthDetail.DvAuthKey,
				DvAuthValue: *resp.Response.DvAuthDetail.DvAuthValue,
			},
		}, nil
	}
	return nil, nil
}

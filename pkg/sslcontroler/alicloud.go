package sslcontroler

import (
	"fmt"

	alidns20150109 "github.com/alibabacloud-go/alidns-20150109/v4/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	"github.com/alibabacloud-go/tea/tea"
	credential "github.com/aliyun/credentials-go/credentials"
)

type ALiCloudService struct {
	client *alidns20150109.Client
	domain string
}

type ALiCloudServiceOption struct {
	AccessKeyId     string
	AccessKeySecret string
	Domain          string
}

func NewALiCloudService(option *ALiCloudServiceOption) (*ALiCloudService, error) {
	if option.AccessKeyId == "" || option.AccessKeySecret == "" || option.Domain == "" {
		return nil, fmt.Errorf("ALiCloudService Init failed: parameter missing")
	}
	access_key := "access_key"
	credential, err := credential.NewCredential(&credential.Config{
		Type:            &access_key,
		AccessKeyId:     &option.AccessKeyId,
		AccessKeySecret: &option.AccessKeySecret,
	})
	if err != nil {
		return nil, err
	}

	config := &openapi.Config{
		Credential: credential,
	}
	config.Endpoint = tea.String("alidns.aliyuncs.com")
	client, err := alidns20150109.NewClient(config)
	if err != nil {
		return nil, err
	}

	if option.AccessKeyId == "" || option.AccessKeySecret == "" || option.Domain == "" {
		return nil, fmt.Errorf("ALiCloudService Init failed: parameter missing")
	}
	t := &ALiCloudService{}
	t.domain = option.Domain
	t.client = client
	return t, nil
}

func (s *ALiCloudService) AddDomainRecord(rr, tp, dv string) (string, error) {
	if dv == "" {
		return "", fmt.Errorf("AddDomainRecord failed: dv parameter missing")
	}
	if rr == "" {
		rr = "_dnsauth"
	}
	if tp == "" {
		tp = "TXT"
	}
	request := &alidns20150109.AddDomainRecordRequest{
		DomainName: tea.String(s.domain),
		RR:         tea.String(rr),
		Type:       tea.String(tp),
		Value:      tea.String(dv),
	}
	response, err := s.client.AddDomainRecord(request)
	if *response.StatusCode != 200 {
		return "", err
	}
	return *response.Body.RecordId, nil
}

func (s *ALiCloudService) ApplyCertificate() (string, error) {
	// ALiCloud does not support applying SSL certificates via API
	return "", fmt.Errorf("ALiCloudService ApplyCertificate not implemented")
}

func (s *ALiCloudService) GetCertificate(id string) (*Certificate, *CertificateKey, error) {
	// ALiCloud does not support downloading SSL certificates via API
	return nil, nil, fmt.Errorf("ALiCloudService GetCertificate not implemented")
}

func (s *ALiCloudService) CheckCertificate(id string) (bool, error) {
	// ALiCloud does not support checking SSL certificate status via API
	return false, fmt.Errorf("ALiCloudService CheckCertificate not implemented")
}

func (s *ALiCloudService) GetDVAuthDetail(id string) (string, error) {
	// ALiCloud does not support getting DV auth details via API
	return "", fmt.Errorf("ALiCloudService GetDVAuthDetail not implemented")
}

func (s *ALiCloudService) CheckDVAuthStatus(id string) (*CertInfo, error) {
	// ALiCloud does not support checking DV auth status via API
	return nil, fmt.Errorf("ALiCloudService CheckDVAuthStatus not implemented")
}

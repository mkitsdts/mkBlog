package sslcontroler

// public certificate
type Certificate struct {
	Data *[]byte
	Name string
}

// private key
type CertificateKey struct {
	Data *[]byte
	Name string
}

type DvAuthDetail struct {
	DvAuthKey   string // DNS record name
	DvAuthValue string // DNS record content
}

type CertInfo struct {
	Status      int    // 0:under review 1:approved 2:rejected
	CertEndTime string // certificate expiration time
	DvDetail    *DvAuthDetail
}

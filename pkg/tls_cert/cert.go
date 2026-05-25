package tlscert

import (
	"crypto/tls"
	"errors"
	"mkBlog/config"
	"sync"
)

var currentCert *tls.Certificate
var certMux sync.RWMutex

func LoadCert() error {
	newCert, err := tls.LoadX509KeyPair(config.Cfg.TLS.Cert, config.Cfg.TLS.Key)
	if err != nil {
		return err
	}
	certMux.Lock()
	currentCert = &newCert
	certMux.Unlock()
	return nil
}

func GetCurrentCert(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	certMux.RLock()
	defer certMux.RUnlock()
	if currentCert == nil {
		return nil, errors.New("no certificate loaded")
	}
	return currentCert, nil
}

func updateCert() error {
	newCert, err := tls.LoadX509KeyPair(config.Cfg.TLS.Cert, config.Cfg.TLS.Key)
	if err != nil {
		return err
	}
	certMux.Lock()
	currentCert = &newCert
	certMux.Unlock()
	return nil
}

package tlscert

import (
	"crypto/tls"
	"log/slog"
	"mkBlog/config"
	"sync"
)

var currentCert *tls.Certificate
var certMux sync.RWMutex

func LoadCert() {
	newCert, err := tls.LoadX509KeyPair(config.Cfg.TLS.Cert, config.Cfg.TLS.Key)
	if err != nil {
		slog.Error("Failed to load X509 certfile.", " check error: ", err)
	}
	certMux.Lock()
	defer certMux.Unlock()
	currentCert = &newCert
}

func GetCurrentCert(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	certMux.RLock()
	defer certMux.RUnlock()
	return currentCert, nil
}

func updateCert() {
	newCert, err := tls.LoadX509KeyPair(config.Cfg.TLS.Cert, config.Cfg.TLS.Key)
	if err != nil {
		slog.Error("Failed to load X509 certfile.", " check error: ", err)
	}
	certMux.Lock()
	defer certMux.Unlock()
	currentCert = &newCert
}

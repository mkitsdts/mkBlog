package tlscert

import (
	"crypto/tls"
	"errors"
	"log/slog"
	"mkBlog/config"
	"sync"
	"time"
)

// "证书未准备"日志的频率
const certNotLoadedWarnInterval = time.Minute

var (
	currentCert *tls.Certificate
	certMux     sync.RWMutex

	lastWarnMux sync.Mutex
	lastWarnAt  time.Time
)

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

// 获取当前证书，证书没准备时返回错误，让握手失败但进程继续运行，等待 cert_control 签发
func GetCurrentCert(_ *tls.ClientHelloInfo) (*tls.Certificate, error) {
	certMux.RLock()
	cert := currentCert
	certMux.RUnlock()
	if cert == nil {
		warnCertNotLoaded()
		return nil, errors.New("TLS certificate is not loaded yet")
	}
	return cert, nil
}

func warnCertNotLoaded() {
	lastWarnMux.Lock()
	defer lastWarnMux.Unlock()
	if time.Since(lastWarnAt) < certNotLoadedWarnInterval {
		return
	}
	lastWarnAt = time.Now()
	slog.Warn("TLS certificate is not loaded yet, handshakes will fail until it is issued",
		"cert", config.Cfg.TLS.Cert, "key", config.Cfg.TLS.Key)
}

func updateCert() error {
	return LoadCert()
}

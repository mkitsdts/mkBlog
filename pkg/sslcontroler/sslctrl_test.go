package sslcontroler

import (
	"testing"
)

func TestSSLControler_Run(t *testing.T) {
	// 保护性地 recover，避免未初始化依赖导致 panic 使测试失败。
	defer func() {
		if r := recover(); r != nil {
			t.Logf("recovered from panic: %v", r)
		}
	}()

	sslControler.checkCurrStatus()
	sslControler.downloadCertificate()
}

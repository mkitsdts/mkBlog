package tlscert

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

// 提前续期的时间窗口，证书剩余有效期小于 renewBefore 时开始续期
const renewBefore = 14 * 24 * time.Hour

// 是磁盘上证书文件的状态
type certState int

const (
	// 证书可用且不在续期窗口内
	CertValid certState = iota

	// 证书可用，但剩余有效期不足 renewBefore
	CertExpiring

	// 证书文件不存在，需要签发
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

// 证书检查结果
type certInfo struct {
	State    certState
	NotAfter time.Time
	Domains  []string
}

// 检查证书与私钥文件的状态
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

	// 证书和私钥必须成对可用，否则服务起不来；
	// 这种情况要人工介入，自动重新签发只会在错误的路径上反复烧配额
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

const stateFileName = "state.json"

// 记录证书签发过程的时间线，落盘后有两个作用：
//  1. 重启后继续退避，避免"重启一次就多申请一张证书"烧掉签发配额
//  2. 出事之后可以还原现场：上次申请是什么时候、为什么、连续失败几次
type attemptState struct {
	LastAttemptAt    time.Time `json:"last_attempt_at,omitzero"`
	LastSuccessAt    time.Time `json:"last_success_at,omitzero"`
	LastCertNotAfter time.Time `json:"last_cert_not_after,omitzero"`
	LastReason       string    `json:"last_reason,omitempty"`
	ConsecutiveFails int       `json:"consecutive_failures"`
	LastError        string    `json:"last_error,omitempty"`
}

func statePath() string {
	return filepath.Join(acmeDir(), stateFileName)
}

// 读取落盘申请记录
func loadAttemptState() attemptState {
	raw, err := os.ReadFile(statePath())
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			// 读不出来落盘申请记录时告警，按零值处理
			slog.Warn("failed to read certificate issuance state", "path", statePath(), "error", err)
		}
		return attemptState{}
	}

	var state attemptState
	if err := json.Unmarshal(raw, &state); err != nil {
		slog.Warn("failed to parse certificate issuance state", "path", statePath(), "error", err)
		return attemptState{}
	}
	return state
}

// 落盘当前状态
func (s attemptState) save() attemptState {
	raw, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		// 失败告警
		slog.Warn("failed to encode certificate issuance state", "error", err)
		return s
	}
	if err := writeFileAtomic(statePath(), append(raw, '\n'), 0600); err != nil {
		// 状态文件写不了不应该阻止证书续期
		slog.Warn("failed to save certificate issuance state", "path", statePath(), "error", err)
	}
	return s
}

// 返回距离下一次允许申请还要等多久
func (s attemptState) wait(now time.Time) time.Duration {
	if s.LastAttemptAt.IsZero() {
		return 0
	}
	remaining := s.LastAttemptAt.Add(retryDelay(s.ConsecutiveFails)).Sub(now)
	if remaining <= 0 {
		return 0
	}
	return remaining
}

// 真正发起申请之前把时间点落盘。
func (s attemptState) attempted(info certInfo, now time.Time) attemptState {
	s.LastAttemptAt = now
	s.LastReason = info.State.String()
	s.LastCertNotAfter = info.NotAfter
	return s
}

func (s attemptState) failed(err error, now time.Time) attemptState {
	// 失败本身也是一次尝试：必须把时间戳推进到当前时间，否则按旧的 LastAttemptAt 算退避会得到负数，直接变成不停重试的死循环
	s.LastAttemptAt = now
	s.ConsecutiveFails++
	s.LastError = err.Error()
	return s
}

func (s attemptState) healthy(info certInfo) attemptState {
	s.ConsecutiveFails = 0
	s.LastError = ""
	s.LastCertNotAfter = info.NotAfter
	return s
}

func (s attemptState) succeeded(info certInfo, now time.Time) attemptState {
	s = s.healthy(info)
	s.LastAttemptAt = now
	s.LastSuccessAt = now
	s.LastReason = ""
	return s
}

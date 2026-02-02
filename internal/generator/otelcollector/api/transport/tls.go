package transport

import "time"

type TLSVersion string

const (
	TLSVersionTLS10 TLSVersion = "1.0"
	TLSVersionTLS11 TLSVersion = "1.1"
	TLSVersionTLS12 TLSVersion = "1.2"
	TLSVersionTLS13 TLSVersion = "1.3"
)

type TLS struct {
	Insecure           bool          `json:"insecure,omitempty"`
	InsecureSkipVerify bool          `json:"insecure_skip_verify,omitempty"`
	CertFile           string        `json:"cert_file,omitempty"`
	KeyFile            string        `json:"key_file,omitempty"`
	CAFile             string        `json:"ca_file,omitempty"`
	MinVersion         TLSVersion    `json:"min_version,omitempty"`
	MaxVersion         TLSVersion    `json:"max_version,omitempty"`
	CipherSuites       []string      `json:"cipher_suites,omitempty"`
	ReloadInterval     time.Duration `json:"reload_interval,omitempty"`
}

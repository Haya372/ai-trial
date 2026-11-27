package semgrepcanary

import "crypto/tls"

func Config() *tls.Config {
	return &tls.Config{InsecureSkipVerify: true}
}

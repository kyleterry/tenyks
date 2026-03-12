package certutil

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"fmt"
	"strings"
)

// OIDTenyksPerms is the X.509 extension OID used to encode allowed destination
// paths in a tenyks client certificate.
//
// Arc: 1.3.6.1.4.1.57036 (tenyks private enterprise number)
//
//	.1   permission extensions
//	.1   allowed paths
var OIDTenyksPerms = asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 57036, 1, 1}

// Permissions holds the access control rules encoded in a client certificate.
type Permissions struct {
	// Paths is the list of destination paths this service may receive messages
	// from. Each entry is matched as an exact path ("libera/#general") or a
	// prefix ("libera" matches any channel on that server). An empty list means
	// all paths are allowed.
	Paths []string
}

// AllowsPath reports whether the permissions allow receiving a message with the
// given destination path. An empty path list means no restriction.
func (p Permissions) AllowsPath(destPath string) bool {
	if len(p.Paths) == 0 {
		return true
	}
	for _, allowed := range p.Paths {
		if allowed == destPath {
			return true
		}
		// "libera" matches "libera/#general", "libera/#random", etc.
		if strings.HasPrefix(destPath, allowed+"/") {
			return true
		}
	}
	return false
}

// EncodeExtensions encodes p as a slice of pkix.Extension values ready to
// embed in a certificate. Returns an empty slice when p has no restrictions.
func EncodeExtensions(p Permissions) ([]pkix.Extension, error) {
	if len(p.Paths) == 0 {
		return nil, nil
	}
	val, err := asn1.Marshal(strings.Join(p.Paths, ","))
	if err != nil {
		return nil, fmt.Errorf("encode permissions: %w", err)
	}
	return []pkix.Extension{{Id: OIDTenyksPerms, Value: val}}, nil
}

// DecodePerms extracts Permissions from cert. If no tenyks extension is
// present, a zero Permissions is returned (allow all).
func DecodePerms(cert *x509.Certificate) (Permissions, error) {
	for _, ext := range cert.Extensions {
		if !ext.Id.Equal(OIDTenyksPerms) {
			continue
		}
		var s string
		if _, err := asn1.Unmarshal(ext.Value, &s); err != nil {
			return Permissions{}, fmt.Errorf("decode permissions: %w", err)
		}
		if s == "" {
			return Permissions{}, nil
		}
		return Permissions{Paths: strings.Split(s, ",")}, nil
	}
	return Permissions{}, nil
}

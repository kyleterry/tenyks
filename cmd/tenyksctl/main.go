package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"filippo.io/age"
	"github.com/kyleterry/tenyks/internal/certutil"
)

const usage = `tenyksctl - tenyks administration tool

Usage:
  tenyksctl <command> [flags]

Commands:
  generate-client-certificate    Generate a client certificate with encoded permissions

Run "tenyksctl <command> -help" for command-specific flags.
`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "generate-client-certificate":
		runGenerateClientCertificate(os.Args[2:])
	case "-help", "--help", "-h":
		fmt.Fprint(os.Stderr, usage)
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n%s", os.Args[1], usage)
		os.Exit(1)
	}
}

func runGenerateClientCertificate(args []string) {
	fs := flag.NewFlagSet("generate-client-certificate", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, `Usage: tenyksctl generate-client-certificate [flags]

Generate a new mTLS client certificate signed by the CA. Permissions are
encoded as custom X.509 extensions. Use -bundle to produce an age-encrypted
archive safe to deliver over an insecure channel.

Flags:`)
		fs.PrintDefaults()
	}

	caCert := fs.String("ca-cert", "", "path to CA certificate file (required)")
	caKey := fs.String("ca-key", "", "path to CA private key file (required)")
	name := fs.String("name", "", "service name — used as the certificate CN (required)")
	paths := fs.String("paths", "", `comma-separated list of allowed destination paths (empty = all).
    Examples:
      libera/#general        exact server+channel
      libera                 all channels on that server
      #general               that channel on any server`)
	bundle := fs.Bool("bundle", false, "produce an age-encrypted bundle instead of writing cert/key to disk")
	agePublicKey := fs.String("age-public-key", "", "recipient age public key (required with -bundle)")
	outBundle := fs.String("out", "", "output path for the encrypted bundle (default: <name>.tar.gz.age, only used with -bundle)")
	outCert := fs.String("out-cert", "", "output path for certificate PEM (default: <name>.crt, only used without -bundle)")
	outKey := fs.String("out-key", "", "output path for private key PEM (default: <name>.key, only used without -bundle)")
	days := fs.Int("days", 365, "certificate validity in days")

	_ = fs.Parse(args)

	var missing []string
	if *caCert == "" {
		missing = append(missing, "-ca-cert")
	}
	if *caKey == "" {
		missing = append(missing, "-ca-key")
	}
	if *name == "" {
		missing = append(missing, "-name")
	}
	if *bundle && *agePublicKey == "" {
		missing = append(missing, "-age-public-key (required with -bundle)")
	}
	if len(missing) > 0 {
		fmt.Fprintf(os.Stderr, "generate-client-certificate: missing required flags: %s\n", strings.Join(missing, ", "))
		os.Exit(1)
	}

	var perms certutil.Permissions
	if *paths != "" {
		for p := range strings.SplitSeq(*paths, ",") {
			p = strings.TrimSpace(p)
			if p != "" {
				perms.Paths = append(perms.Paths, p)
			}
		}
	}

	certPEM, keyPEM, err := certutil.GenerateClientCert(*name, *caCert, *caKey, perms, *days)
	if err != nil {
		fmt.Fprintf(os.Stderr, "generate-client-certificate: %v\n", err)
		os.Exit(1)
	}

	if *bundle {
		if *outBundle == "" {
			*outBundle = *name + ".tar.gz.age"
		}

		recipient, err := age.ParseX25519Recipient(*agePublicKey)
		if err != nil {
			fmt.Fprintf(os.Stderr, "generate-client-certificate: invalid age public key: %v\n", err)
			os.Exit(1)
		}

		caPEM, err := os.ReadFile(*caCert)
		if err != nil {
			fmt.Fprintf(os.Stderr, "generate-client-certificate: read ca cert: %v\n", err)
			os.Exit(1)
		}

		tarGzData, err := makeTarGz(*name, certPEM, keyPEM, caPEM)
		if err != nil {
			fmt.Fprintf(os.Stderr, "generate-client-certificate: create bundle: %v\n", err)
			os.Exit(1)
		}

		f, err := os.OpenFile(*outBundle, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "generate-client-certificate: create bundle file: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()

		w, err := age.Encrypt(f, recipient)
		if err != nil {
			fmt.Fprintf(os.Stderr, "generate-client-certificate: age encrypt: %v\n", err)
			os.Exit(1)
		}

		if _, err := io.Copy(w, bytes.NewReader(tarGzData)); err != nil {
			fmt.Fprintf(os.Stderr, "generate-client-certificate: encrypt bundle: %v\n", err)
			os.Exit(1)
		}

		if err := w.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "generate-client-certificate: finalize bundle: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("bundle      : %s\n", *outBundle)
	} else {
		if *outCert == "" {
			*outCert = *name + ".crt"
		}
		if *outKey == "" {
			*outKey = *name + ".key"
		}

		if err := os.WriteFile(*outCert, certPEM, 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "generate-client-certificate: write cert: %v\n", err)
			os.Exit(1)
		}
		if err := os.WriteFile(*outKey, keyPEM, 0o600); err != nil {
			fmt.Fprintf(os.Stderr, "generate-client-certificate: write key: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("certificate : %s\n", *outCert)
		fmt.Printf("private key : %s\n", *outKey)
	}

	fmt.Printf("common name : %s\n", *name)
	fmt.Printf("valid days  : %d\n", *days)
	if len(perms.Paths) == 0 {
		fmt.Println("paths       : (all)")
	} else {
		fmt.Printf("paths       : %s\n", strings.Join(perms.Paths, ", "))
	}
}

// makeTarGz packs the cert, key, and CA cert into an in-memory tar.gz archive.
func makeTarGz(name string, certPEM, keyPEM, caPEM []byte) ([]byte, error) {
	files := []struct {
		name string
		mode int64
		data []byte
	}{
		{name + ".crt", 0o644, certPEM},
		{name + ".key", 0o600, keyPEM},
		{"ca.crt", 0o644, caPEM},
	}

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)

	for _, f := range files {
		hdr := &tar.Header{
			Name: f.name,
			Mode: f.mode,
			Size: int64(len(f.data)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return nil, err
		}
		if _, err := tw.Write(f.data); err != nil {
			return nil, err
		}
	}

	if err := tw.Close(); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

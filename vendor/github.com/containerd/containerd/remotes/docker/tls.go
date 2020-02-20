package docker

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/go-connections/tlsconfig"
)

var CertsDir = "/etc/docker/certs.d"

/*func newTLSConfig(hostname string, isSecure bool) (*tls.Config, error) {
	// PreferredServerCipherSuites should have no effect
	tlsConfig := tlsconfig.ServerDefault()

	tlsConfig.InsecureSkipVerify = !isSecure

	if isSecure && CertsDir != "" {
		hostDir := filepath.Join(CertsDir, hostname)
		if err := readCertsDirectory(tlsConfig, hostDir); err != nil {
			return nil, err
		}
	}

	return tlsConfig, nil
}*/

func loadCAs(hostname string, tlsConfig *tls.Config) error {
	hostDir := filepath.Join(CertsDir, hostname)
	return readCertsDirectory(tlsConfig, hostDir)
}

func hasFile(files []os.FileInfo, name string) bool {
	for _, f := range files {
		if f.Name() == name {
			return true
		}
	}
	return false
}

// ReadCertsDirectory reads the directory for TLS certificates
// including roots and certificate pairs and updates the
// provided TLS configuration.
func readCertsDirectory(tlsConfig *tls.Config, directory string) error {
	fs, err := ioutil.ReadDir(directory)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	for _, f := range fs {
		if strings.HasSuffix(f.Name(), ".crt") {
			if tlsConfig.RootCAs == nil {
				systemPool, err := tlsconfig.SystemCertPool()
				if err != nil {
					return fmt.Errorf("unable to get system cert pool: %v", err)
				}
				tlsConfig.RootCAs = systemPool
			}
			data, err := ioutil.ReadFile(filepath.Join(directory, f.Name()))
			if err != nil {
				return err
			}
			tlsConfig.RootCAs.AppendCertsFromPEM(data)
		}
		if strings.HasSuffix(f.Name(), ".cert") {
			certName := f.Name()
			keyName := certName[:len(certName)-5] + ".key"
			if !hasFile(fs, keyName) {
				return fmt.Errorf("missing key %s for client certificate %s. Note that CA certificates should use the extension .crt", keyName, certName)
			}
			cert, err := tls.LoadX509KeyPair(filepath.Join(directory, certName), filepath.Join(directory, keyName))
			if err != nil {
				return err
			}
			tlsConfig.Certificates = append(tlsConfig.Certificates, cert)
		}
		if strings.HasSuffix(f.Name(), ".key") {
			keyName := f.Name()
			certName := keyName[:len(keyName)-4] + ".cert"
			if !hasFile(fs, certName) {
				return fmt.Errorf("Missing client certificate %s for key %s", certName, keyName)
			}
		}
	}

	return nil
}

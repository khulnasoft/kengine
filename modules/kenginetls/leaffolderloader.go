// Copyright 2015 Matthew Holt and The Kengine Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package kenginetls

import (
	"crypto/x509"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/khulnasoft/kengine/v2"
)

func init() {
	kengine.RegisterModule(LeafFolderLoader{})
}

// LeafFolderLoader loads certificates and their associated keys from disk
// by recursively walking the specified directories, looking for PEM
// files which contain both a certificate and a key.
type LeafFolderLoader struct {
	Folders []string `json:"folders,omitempty"`
}

// KengineModule returns the Kengine module information.
func (LeafFolderLoader) KengineModule() kengine.ModuleInfo {
	return kengine.ModuleInfo{
		ID:  "tls.leaf_cert_loader.folder",
		New: func() kengine.Module { return new(LeafFolderLoader) },
	}
}

// Provision implements kengine.Provisioner.
func (fl *LeafFolderLoader) Provision(ctx kengine.Context) error {
	repl, ok := ctx.Value(kengine.ReplacerCtxKey).(*kengine.Replacer)
	if !ok {
		repl = kengine.NewReplacer()
	}
	for k, path := range fl.Folders {
		fl.Folders[k] = repl.ReplaceKnown(path, "")
	}
	return nil
}

// LoadLeafCertificates loads all the leaf certificates in the directories
// listed in fl from all files ending with .pem.
func (fl LeafFolderLoader) LoadLeafCertificates() ([]*x509.Certificate, error) {
	var certs []*x509.Certificate
	for _, dir := range fl.Folders {
		err := filepath.Walk(dir, func(fpath string, info os.FileInfo, err error) error {
			if err != nil {
				return fmt.Errorf("unable to traverse into path: %s", fpath)
			}
			if info.IsDir() {
				return nil
			}
			if !strings.HasSuffix(strings.ToLower(info.Name()), ".pem") {
				return nil
			}

			certData, err := convertPEMFilesToDERBytes(fpath)
			if err != nil {
				return err
			}
			cert, err := x509.ParseCertificate(certData)
			if err != nil {
				return fmt.Errorf("%s: %w", fpath, err)
			}

			certs = append(certs, cert)

			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return certs, nil
}

var (
	_ LeafCertificateLoader = (*LeafFolderLoader)(nil)
	_ kengine.Provisioner   = (*LeafFolderLoader)(nil)
)

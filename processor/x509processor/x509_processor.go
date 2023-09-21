// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package x509processor // import "github.com/open-telemetry/opentelemetry-collector-contrib/processor/x509processor"

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"os"

	"go.opentelemetry.io/collector/pdata/plog"
	"go.uber.org/zap"
)

type x509Processor struct {
	logger *zap.Logger

	keyID      string
	privateKey *rsa.PrivateKey

	attributeOperation map[string]Operation
}

func (self *x509Processor) loadOperation(cfg *Config) {
	self.attributeOperation = make(map[string]Operation)
	for _, a := range cfg.SignAttribute {
		self.attributeOperation[a] = Sign
	}

	for _, a := range cfg.EncryptAttribute {
		self.attributeOperation[a] = Encrypt
	}
}

func (self *x509Processor) loadKey(filename string) error {
	privateKeyFile, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	privateKeyBlock, _ := pem.Decode(privateKeyFile)

	hash := sha256.New()
	hash.Write(privateKeyBlock.Bytes)
	self.keyID = fmt.Sprintf("%x", hash.Sum(nil))

	key, err := x509.ParsePKCS8PrivateKey(privateKeyBlock.Bytes)
	if err != nil {
		return err
	}

	if _, ok := key.(*rsa.PrivateKey); !ok {
		return errors.New("only rsa private keys are supported")
	}

	self.privateKey = key.(*rsa.PrivateKey)
	self.logger.Info("loaded key for encryption and signing", zap.String("key id", self.keyID))
	return nil
}

func (self *x509Processor) signValue(ctx context.Context, k string, v any) string {

	hash := sha256.New()
	hash.Write([]byte(fmt.Sprintf("%v", v)))
	digest := hash.Sum(nil)

	signature, err := rsa.SignPKCS1v15(rand.Reader, self.privateKey, crypto.SHA256, digest)
	if err != nil {
		self.logger.Error("unable to sign attribute", zap.String("attribute", k))
		return ""
	}

	return base64.StdEncoding.EncodeToString(signature)
}

func (self *x509Processor) encryptValue(ctx context.Context, k string, v any) string {

	data, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, &self.privateKey.PublicKey, []byte(fmt.Sprintf("%v", v)), []byte{})
	if err != nil {
		self.logger.Error("unable to encrypt attribute", zap.String("attribute", k))
		return ""
	}
	return base64.StdEncoding.EncodeToString(data)
}

func (self *x509Processor) processLogs(ctx context.Context, ld plog.Logs) (plog.Logs, error) {
	rls := ld.ResourceLogs()

	for i := 0; i < rls.Len(); i++ {
		attr := rls.At(i).Resource().Attributes()

		opCount := 0

		for k, v := range attr.AsRaw() {
			op, ok := self.attributeOperation[k]
			if !ok {
				continue
			}

			opCount++

			switch op {
			case Sign:
				attr.PutStr(fmt.Sprintf("_x509_sig_%s", k), self.signValue(ctx, k, v))
			case Encrypt:
				attr.Remove(k) // remove original value
				attr.PutStr(fmt.Sprintf("_x509_enc_%s", k), self.encryptValue(ctx, k, v))
			}
		}

		if opCount > 0 {
			attr.PutStr("_x509_key", self.keyID) // add the hash of the key so we know which one to use during validation/decrypt
		}

	}

	return ld, nil
}

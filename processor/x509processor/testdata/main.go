package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"os"
)

func main() {

	// Sign

	privateKeyFile, err := os.ReadFile("../key.pem")
	if err != nil {
		fmt.Println("Error loading private key file:", err)
		os.Exit(1)
	}

	privateKeyBlock, _ := pem.Decode(privateKeyFile)

	hash := sha256.New()
	hash.Write(privateKeyBlock.Bytes)
	keyID := hash.Sum(nil)

	fmt.Printf("keyID: %x\n", keyID)

	key, err := x509.ParsePKCS8PrivateKey(privateKeyBlock.Bytes)
	if err != nil {
		fmt.Println("Error parsing private key:", err)
		os.Exit(1)
	}
	privateKey := key.(*rsa.PrivateKey)

	message := []byte("confidential log message")
	hash = sha256.New()
	hash.Write(message)
	digest := hash.Sum(nil)

	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, digest)
	if err != nil {
		fmt.Println("Error signing data:", err)
		os.Exit(1)
	}

	fmt.Println("Signature:", base64.StdEncoding.EncodeToString(signature))

	// Verify signature

	if err := rsa.VerifyPKCS1v15(&privateKey.PublicKey, crypto.SHA256, digest, signature); err != nil {
		fmt.Println("Verfication error:", err)
		os.Exit(1)
	}
	fmt.Printf("Signature valid\n\n")

	// Encrypt

	data, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, &privateKey.PublicKey, message, []byte{})
	if err != nil {
		fmt.Println("Encryption error:", err)
		os.Exit(1)
	}
	fmt.Println("Encrypted:", base64.StdEncoding.EncodeToString(data))

	// Decrypt

	decrypted, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, data, []byte{})
	if err != nil {
		fmt.Println("Decryption error:", err)
		os.Exit(1)
	}
	fmt.Println("Decrypted:", string(decrypted))

}

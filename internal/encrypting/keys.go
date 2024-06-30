package encrypting

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io"
	"os"
	"path"
)

const (
	privateKeyName = "private.pem"
	publicKeyName  = "public.pem"
	KeyLengthBits  = 4096
)

type keypair struct {
	keysPath string
}

func NewService(keyPath string) *keypair {
	return &keypair{
		keysPath: keyPath,
	}
}

func (k *keypair) GenerateIfNotExist() error {
	privateKeyPath := path.Join(k.keysPath, privateKeyName)
	publicKeyPath := path.Join(k.keysPath, publicKeyName)

	privateKeyFile, err := os.OpenFile(privateKeyPath, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer privateKeyFile.Close()

	publicKeyFile, err := os.OpenFile(publicKeyPath, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer publicKeyFile.Close()

	privateKey, err := rsa.GenerateKey(rand.Reader, KeyLengthBits)
	if err != nil {
		return err
	}

	var privateKeyPEM, publicKeyPEM bytes.Buffer

	err = pem.Encode(&privateKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})
	if err != nil {
		return err
	}

	err = pem.Encode(&publicKeyPEM, &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(&privateKey.PublicKey),
	})
	if err != nil {
		return err
	}

	_, err = privateKeyFile.Write(privateKeyPEM.Bytes())
	if err != nil {
		return err
	}

	_, err = publicKeyFile.Write(publicKeyPEM.Bytes())
	if err != nil {
		return err
	}

	return nil
}

func (k *keypair) GetPrivateKey() (*rsa.PrivateKey, error) {
	privateKeyPath := path.Join(k.keysPath, privateKeyName)

	privateKeyFile, err := os.Open(privateKeyPath)
	if err != nil {
		return nil, err
	}
	defer privateKeyFile.Close()

	privateKeyBytes, err := io.ReadAll(privateKeyFile)
	if err != nil {
		return nil, err
	}

	privateKeyBlock, _ := pem.Decode(privateKeyBytes)

	privateKey, err := x509.ParsePKCS1PrivateKey(privateKeyBlock.Bytes)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

func (k *keypair) GetPublicKey() (*rsa.PublicKey, error) {
	publicKeyPath := path.Join(k.keysPath, publicKeyName)

	publicKeyFile, err := os.Open(publicKeyPath)
	if err != nil {
		return nil, err
	}
	defer publicKeyFile.Close()

	publicKeyBytes, err := io.ReadAll(publicKeyFile)
	if err != nil {
		return nil, err
	}

	publicKeyBlock, _ := pem.Decode(publicKeyBytes)

	publicKey, err := x509.ParsePKCS1PublicKey(publicKeyBlock.Bytes)
	if err != nil {
		return nil, err
	}

	return publicKey, nil
}

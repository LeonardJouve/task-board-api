package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"log"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

func getRSAKeys(name string) ([]byte, []byte, error) {
	privateFilename, publicFilename := getRSAFilenames(name)

	_, privateErr := os.Stat(privateFilename)
	_, publicErr := os.Stat(publicFilename)
	if privateErr != nil || publicErr != nil {
		return generateKeys(name)
	}

	privatePEM, privateErr := os.ReadFile(privateFilename)
	publicPEM, publicErr := os.ReadFile(publicFilename)
	if privateErr != nil || publicErr != nil {
		return generateKeys(name)
	}

	privateKey, privateErr := jwt.ParseRSAPrivateKeyFromPEM(privatePEM)
	publicKey, publicErr := jwt.ParseRSAPublicKeyFromPEM(publicPEM)
	if privateErr != nil || publicErr != nil {
		return generateKeys(name)
	}

	if privateKey.PublicKey.N.Cmp(publicKey.N) != 0 {
		return generateKeys(name)
	}

	return privatePEM, publicPEM, nil
}

func generateKeys(name string) ([]byte, []byte, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, err
	}

	privatePEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
		},
	)

	publicPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: x509.MarshalPKCS1PublicKey(&privateKey.PublicKey),
		},
	)

	privateFilename, publicFilename := getRSAFilenames(name)
	if err := os.WriteFile(privateFilename, privatePEM, 0700); err != nil {
		return nil, nil, err
	}
	if err := os.WriteFile(publicFilename, publicPEM, 0755); err != nil {
		return nil, nil, err
	}

	log.Println("GENERATING RSA CERTIFICATES")

	return privatePEM, publicPEM, nil
}

func getRSAFilenames(name string) (string, string) {
	return "rsa/" + name + ".rsa", "rsa/" + name + ".rsa.pub"
}

package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"log"
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func getKeyPEM(c *fiber.Ctx, name string, private bool) ([]byte, bool) {
	filename, err := getRSAFilepath(name, private)
	if err != nil {
		c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "server error",
		})
		return nil, false
	}

	var keyPEM []byte

	if _, err := os.Stat(filename); err != nil {
		_, keyPEM, err = generateKeys(name)
		if err != nil {
			c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "server error",
			})
			return nil, false
		}
	} else {
		keyPEM, err = os.ReadFile(filename)
		if err != nil {
			c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "server error",
			})
			return nil, false
		}
	}

	return keyPEM, true
}

func getPrivateKey(c *fiber.Ctx, name string) (*rsa.PrivateKey, bool) {
	privatePEM, ok := getKeyPEM(c, name, true)
	if !ok {
		return nil, false
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privatePEM)
	if err != nil {
		c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "server error",
		})
		return nil, false
	}

	return privateKey, true
}

func getPublicKey(c *fiber.Ctx, name string) (*rsa.PublicKey, bool) {
	publicPEM, ok := getKeyPEM(c, name, false)
	if !ok {
		return nil, false
	}

	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(publicPEM)
	if err != nil {
		c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "server error",
		})
		return nil, false
	}

	return publicKey, true
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

	publicFilename, err := getRSAFilepath(name, false)
	if err != nil {
		return nil, nil, err
	}

	privateFilename, err := getRSAFilepath(name, true)
	if err != nil {
		return nil, nil, err
	}

	if err := os.WriteFile(publicFilename, publicPEM, 0755); err != nil {
		return nil, nil, err
	}
	if err := os.WriteFile(privateFilename, privatePEM, 0700); err != nil {
		os.Remove(publicFilename)
		return nil, nil, err
	}

	log.Println("GENERATING RSA CERTIFICATES")

	return publicPEM, privatePEM, nil
}

func getRSAFilepath(name string, private bool) (string, error) {
	path, err := os.Executable()
	if err != nil {
		return "", err
	}

	folderpath := filepath.Join(filepath.Dir(path), "..", "rsa")

	if _, err := os.Stat(folderpath); err != nil {
		if err = os.Mkdir(folderpath, 0755); err != nil {
			return "", err
		}
	}

	filename := name + ".rsa"
	if !private {
		filename += ".pub"
	}

	filepath := filepath.Join(folderpath, filename)

	return filepath, nil
}

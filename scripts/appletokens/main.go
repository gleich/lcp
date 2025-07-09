package main

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.mattglei.ch/timber"
)

const KEY_FILENAME = "key.p8"

func main() {
	key, err := os.ReadFile(KEY_FILENAME)
	if err != nil {
		timber.Fatal(err, "failed to read key from", KEY_FILENAME)
	}
	teamID := os.Getenv("TEAM_ID")
	if teamID == "" {
		timber.FatalMsg("Please provide team id through environment variable")
	}
	keyID := os.Getenv("KEY_ID")
	if keyID == "" {
		timber.FatalMsg("Please provide key id through environment variable")
	}
	now := time.Now()

	block, _ := pem.Decode(key)
	if block == nil || block.Type != "PRIVATE KEY" {
		timber.FatalMsg("failed to decode PEM block containing private key")
	}

	keyIfc, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		timber.Fatal(err, "parse PKCS#8 private key", err)
	}
	ecKey, ok := keyIfc.(*ecdsa.PrivateKey)
	if !ok {
		timber.Fatal(err, "not an ECDSA private key")
	}

	claims := jwt.MapClaims{
		"iss": teamID,
		"iat": now.Unix(),
		"exp": now.Add(182 * 24 * time.Hour).Unix(), // expires in 6 months
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token.Header["kid"] = keyID
	token.Header["alg"] = "ES256"

	signed, err := token.SignedString(ecKey)
	if err != nil {
		timber.Fatal(err, "failed to sign token")
	}

	fmt.Println(signed)
}

package pgp

import (
	"bytes"
	"fmt"
	"golang.org/x/crypto/openpgp"
	"log"
	"strings"
)

func GetKeyID(publicKey string) []string {
	entityList, err := openpgp.ReadArmoredKeyRing(bytes.NewBufferString(publicKey))
	if err != nil {
		log.Fatal(err)
	}

	var keys []string
	for _, entity := range entityList {
		fingerPrint := entity.PrimaryKey.Fingerprint
		keyID := fingerPrint[len(fingerPrint)-8:]
		value := fmt.Sprintf("%x", keyID)
		keys = append(keys, strings.ToUpper(value))
	}

	return keys
}

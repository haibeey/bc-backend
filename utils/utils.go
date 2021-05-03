package utils

import (
	"bytes"
	"crypto/rand"
	_ "crypto/sha256"
	"encoding/base64"
	"fmt"
	"golang.org/x/crypto/openpgp/armor"
	_ "golang.org/x/crypto/ripemd160"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
	"unicode"

	"golang.org/x/crypto/openpgp"

	"github.com/jchavannes/go-pgp/pgp"

	"github.com/dgrijalva/jwt-go"
)

// GenerateToken generates a jwt token and assign a username to it's claims and return it
func GenerateToken(username string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	/* Create a map to store our claims */
	claims := token.Claims.(jwt.MapClaims)
	/* Set token claims */
	claims["username"] = username
	claims["exp"] = time.Now().Add(time.Hour * 2400).Unix()
	tokenString, err := token.SignedString([]byte(os.Getenv("SECRETKEY")))
	if err != nil {
		log.Fatal("Error in Generating key", err)
		return "", err
	}
	return tokenString, nil
}

// ParseToken parses a jwt token and returns the username in it's claims
func ParseToken(tokenStr string) (string, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("SECRETKEY")), nil
	})
	if err!=nil{
		return "",err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		username := claims["username"].(string)
		return username, nil
	} else {
		return "", err
	}
}

//SpaceStringsBuilder
func SpaceStringsBuilder(str string) string {
	var b strings.Builder
	b.Grow(len(str))
	for _, ch := range str {
		if !unicode.IsSpace(ch) {
			b.WriteRune(ch)
		}
	}
	return b.String()
}

func ContainsOnlyDigit(str string) bool {
	for _, curValue := range str {
		if curValue < 48 || curValue > 57 {
			return false
		}
	}
	return true
}

func PGPEncrypt1(data string, pubKey string) (string, error) {

	keyToUse, err := base64.StdEncoding.DecodeString(pubKey)
	if err != nil {
		return "", err
	}

	entityList, err := openpgp.ReadArmoredKeyRing(strings.NewReader(string(keyToUse)))
	if err != nil {
		return "", err
	}

	// encrypt string
	buf := new(bytes.Buffer)
	w, err := openpgp.Encrypt(buf, entityList, nil, nil, nil)
	if err != nil {
		return "", err
	}
	_, err = w.Write([]byte(data))
	if err != nil {
		return "", err
	}
	err = w.Close()
	if err != nil {
		return "", err
	}
	// Encode to base64
	bytes, err := ioutil.ReadAll(buf)
	if err != nil {
		return "", err
	}
	bytes = []byte(base64.StdEncoding.EncodeToString(bytes))
	if err != nil {
		return "", err
	}
	start := []byte("-----BEGIN PGP MESSAGE-----\n\n")
	start = append(start, bytes...)
	end := []byte("\n-----END PGP MESSAGE-----")
	start = append(start, end...)

	return base64.StdEncoding.EncodeToString(start), nil
}
func PGPEncrypt(data string, pubKey string) (string, error) {
	pubKey = strings.Trim(pubKey, " \n\t\r")
	data = strings.Trim(data, " \n\t\r")
	keyToUse, err := base64.StdEncoding.DecodeString(pubKey)
	if err != nil {
		return "", err
	}
	pubEntity, err := pgp.GetEntity(keyToUse, nil)
	if err != nil {
		return "", err
	}

	encrypted, err := Encrypt(pubEntity, []byte(data))
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(encrypted), nil
}

func Encrypt(entity *openpgp.Entity, message []byte) ([]byte, error) {
	// Create buffer to write output to
	buf := new(bytes.Buffer)

	// Create encoder
	encoderWriter, err := armor.Encode(buf, "PGP MESSAGE", map[string]string{})
	if err != nil {
		return []byte{}, fmt.Errorf("Error creating OpenPGP armor: %v", err)
	}

	// Create encryptor with encoder
	encryptorWriter, err := openpgp.Encrypt(encoderWriter, []*openpgp.Entity{entity}, nil, nil, nil)
	if err != nil {
		return []byte{}, fmt.Errorf("Error creating entity for encryption: %v", err)
	}

	encryptorWriter.Close()
	encoderWriter.Close()

	// Return buffer output - an encoded, encrypted, and compressed message
	a := buf.Bytes()
	// fmt.Println()
	// fmt.Println(string(a))
	// fmt.Println(entity)
	// fmt.Println(string(a))
	return a, nil
}

func GenerateAlphaCode(size int) string {
	b := make([]byte, size)
	rand.Read(b)

	result := ""
	for i := 0; i < size; i++ {
		curValue := b[i]
		for curValue < 48 || curValue > 57 && curValue < 65 || curValue > 90 && curValue < 97 || curValue > 122 {
			if curValue < 48 {
				curValue += 5
			}
			if curValue > 57 {
				curValue += 8
			}
			if curValue > 122 {
				curValue -= 5
			}
		}
		result += string(curValue)
	}
	return result
}

func GenerateIdempotentKey() string {

	return "ba943ff1-ca16-49b2-ba55-1057e70ca5c7"

	// return fmt.Sprintf("%s-%s-%s-%s-%s",
	// 	strings.ToLower(GenerateAlphaCode(8)),
	// 	strings.ToLower(GenerateAlphaCode(4)),
	// 	strings.ToLower(GenerateAlphaCode(4)),
	// 	strings.ToLower(GenerateAlphaCode(4)),
	// 	strings.ToLower(GenerateAlphaCode(12)),
	// )

}

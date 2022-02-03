package windows

type LocalState struct {
	OSCrypt struct {
		EncryptedKey string `json:"encrypted_key"`
	} `json:"os_crypt"`
}

type DataBlob struct {
	cbData uint32
	pbData *byte
}

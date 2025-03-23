package ssm

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"home-ssm/awslib"
	"io"
	"log"

	"github.com/dgraph-io/badger/v4"
)

const (
	defaultKeyId = "aws/ssm"
)

type DataStore struct {
	db *badger.DB
}

/*
AES-256 uses a 32-byte (256-bit) key
AES-128 uses a 16-byte (128-bit) key
AES-192 uses a 24-byte(192-bit) key
*/
func generateAesKey() ([]byte, error) {

	key := make([]byte, 32)
	_, err := io.ReadFull(rand.Reader, key)
	if err != nil {
		return nil, err
	}

	return key, nil
}

func NewDataStore(db *badger.DB) *DataStore {

	return &DataStore{db: db}
}

func (ds *DataStore) initializeDataStore() error {

	err := ds.db.Update(func(txn *badger.Txn) error {

		_, err := txn.Get([]byte(defaultKeyId))
		if err != nil && errors.Is(err, badger.ErrKeyNotFound) {

			log.Println("Generating default key.")

			key, err := generateAesKey()
			if err == nil {
				err = txn.Set([]byte(defaultKeyId), key)
				return err
			}
		}

		return err
	})

	return err
}

func (ds *DataStore) getParameter(key string) (*Parameter, awslib.APIError) {

	var param Parameter
	var apiError awslib.APIError

	err := ds.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {

				apiError = SsmErrorCodes[ErrParameterNotFound]
				return errors.New(apiError.Description)
			}

			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &param)
		})
	})

	if err != nil {
		log.Println("An error occurred.", err)
		if apiError.Code == "" {
			return nil, SsmErrorCodes[ErrInternalError]
		}
		return nil, apiError
	}

	return &param, SsmErrorCodes[ErrNone]
}

func (ds *DataStore) putParameter(key string, value *Parameter, overwrite bool) (int64, awslib.APIError) {

	var apiError awslib.APIError
	var newVersion int64 = -1
	var existingParam Parameter

	err := ds.db.Update(func(txn *badger.Txn) error {

		item, err := txn.Get([]byte(key))

		if err == nil {

			if err := item.Value(func(val []byte) error {
				return json.Unmarshal(val, &existingParam)
			}); err != nil {
				apiError = SsmErrorCodes[ErrInternalError]
				return err
			}

			if !overwrite {
				apiError = SsmErrorCodes[ErrParameterAlreadyExists]
				return errors.New(apiError.Description)
			}

			newVersion = existingParam.Version + 1

		} else if !errors.Is(err, badger.ErrKeyNotFound) {

			apiError = SsmErrorCodes[ErrInternalError]
			return errors.New(apiError.Description)
		}

		value.Version = newVersion
		paramBytes, err := json.Marshal(value)
		if err != nil {
			apiError = SsmErrorCodes[ErrInternalError]
			return err
		}

		return txn.Set([]byte(key), paramBytes)
	})

	if err != nil {
		log.Println("An error occurred.", err)
		if apiError.Code == "" {
			return -1, SsmErrorCodes[ErrInternalError]
		}
		return -1, apiError
	}

	return newVersion, SsmErrorCodes[ErrNone]
}

func (ds *DataStore) findKeyId(keyId string) ([]byte, error) {

	var key []byte
	err := ds.db.View(func(txn *badger.Txn) error {

		if keyId == "" {
			keyId = defaultKeyId
		}

		item, err := txn.Get([]byte(keyId))
		if err != nil {
			return err
		}

		key, err = item.ValueCopy(nil)

		return err
	})

	return key, err
}

func (ds *DataStore) encrypt(stringToEncrypt string, keyId string) (string, error) {

	key, err := ds.findKeyId(keyId)
	if err != nil {
		return "", err
	}

	// Since the key is in string format, convert it to bytes
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// Create a new GCM - https://en.wikipedia.org/wiki/Galois/Counter_Mode
	// https://golang.org/pkg/crypto/cipher/#NewGCM
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Create a nonce. Nonce should never be reused with the same key.
	// Since we use GCM, we recommend using 12 bytes.
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// Encrypt the data using aesGCM.Seal. Since we don't want to save the nonce somewhere else in this case, we add it as a prefix to the encrypted data. The first nonce argument in Seal is the prefix.
	ciphertext := aesGCM.Seal(nonce, nonce, []byte(stringToEncrypt), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (ds *DataStore) decrypt(encryptedString string, keyId string) (string, error) {

	key, err := ds.findKeyId(keyId)
	if err != nil {
		return "", err
	}

	enc, err := base64.StdEncoding.DecodeString(encryptedString)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aesGCM.NonceSize()

	nonce, ciphertext := enc[:nonceSize], enc[nonceSize:]

	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

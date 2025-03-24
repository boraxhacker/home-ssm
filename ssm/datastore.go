package ssm

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"log"
	"strings"

	"github.com/dgraph-io/badger/v4"
)

const (
	DefaultKeyId = "aws/ssm"
)

type DataStore struct {
	db *badger.DB
}

type KeyFilter struct {
	Path       string
	StartsWith bool
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

	err := db.Update(func(txn *badger.Txn) error {

		_, err := txn.Get([]byte(DefaultKeyId))
		if err != nil && errors.Is(err, badger.ErrKeyNotFound) {

			log.Println("Generating default key.")

			key, err := generateAesKey()
			if err == nil {
				err = txn.Set([]byte(DefaultKeyId), key)
				return err
			}
		}

		return err
	})

	if err != nil {

		log.Panicln("Unable to create datastore.", err)
	}

	return &DataStore{db: db}
}

func (ds *DataStore) delete(key string) error {

	err := ds.db.Update(
		func(txn *badger.Txn) error {
			return txn.Delete([]byte(key))
		})

	if err != nil {

		if errors.Is(err, badger.ErrKeyNotFound) {

			return ErrParameterNotFound
		}

		return err
	}

	return nil
}

func (ds *DataStore) findParametersByKey(keyFilters []KeyFilter) ([]Parameter, error) {

	var result []Parameter

	err := ds.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			key := string(item.Key())
			for _, keyFilter := range keyFilters {

				if (keyFilter.Path == key) || (keyFilter.StartsWith && strings.HasPrefix(key, keyFilter.Path)) {

					var param Parameter
					umerr := item.Value(func(val []byte) error {
						return json.Unmarshal(val, &param)
					})

					if umerr == nil {

						result = append(result, param)

					} else {

						return umerr
					}
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (ds *DataStore) getParameter(key string) (*Parameter, error) {

	var param Parameter

	err := ds.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {

				return ErrParameterNotFound
			}

			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &param)
		})
	})

	if err != nil {
		log.Println("An error occurred.", err)
		return nil, err
	}

	return &param, nil
}

func (ds *DataStore) putParameter(key string, value *Parameter, overwrite bool) (int64, error) {

	var newVersion int64 = 1
	var existingParam Parameter

	err := ds.db.Update(func(txn *badger.Txn) error {

		item, err := txn.Get([]byte(key))

		if err == nil {

			if err := item.Value(func(val []byte) error {
				return json.Unmarshal(val, &existingParam)
			}); err != nil {
				return err
			}

			if !overwrite {
				return ErrParameterAlreadyExists
			}

			newVersion = existingParam.Version + 1

		} else if !errors.Is(err, badger.ErrKeyNotFound) {

			return err
		}

		value.Version = newVersion
		paramBytes, err := json.Marshal(value)
		if err != nil {
			return err
		}

		return txn.Set([]byte(key), paramBytes)
	})

	if err != nil {
		return -1, err
	}

	return newVersion, nil
}

func (ds *DataStore) findKeyId(keyId string) ([]byte, error) {

	var key []byte
	err := ds.db.View(func(txn *badger.Txn) error {

		if keyId == "" {
			keyId = DefaultKeyId
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

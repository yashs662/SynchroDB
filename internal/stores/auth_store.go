// internal/stores/auth_store/auth_store.go

package stores

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"golang.org/x/crypto/argon2"
)

// Parameters for Argon2 hashing
const (
	saltLength  = 16 // 128 bits
	timeCost    = 1
	memoryCost  = 64 * 1024 // 64 MB
	parallelism = 1
	keyLength   = 32 // 256 bits
)

// User structure that stores username, hashed password, and role
type User struct {
	Username       string `json:"username"`
	HashedPassword string `json:"hashed_password"`
	Role           string `json:"role"` // admin, read-only, write-only
}

// CredentialStore holds a list of users
type CredentialStore struct {
	Users []User `json:"users"`
}

// HashPassword hashes a password using Argon2id
func HashPassword(password string) (string, error) {
	salt := make([]byte, saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	// Generate the hash using Argon2id
	hash := argon2.IDKey([]byte(password), salt, timeCost, memoryCost, parallelism, keyLength)

	// Combine salt and hash, and encode as base64 for storage
	fullHash := append(salt, hash...)
	encodedHash := base64.StdEncoding.EncodeToString(fullHash)
	return encodedHash, nil
}

// ComparePassword compares a stored hashed password with a provided plaintext password
func ComparePassword(storedHash, password string) error {
	// Decode the stored hash from base64
	fullHash, err := base64.StdEncoding.DecodeString(storedHash)
	if err != nil {
		return err
	}

	// Split the salt and the actual hash
	if len(fullHash) < saltLength+keyLength {
		return errors.New("invalid hash length")
	}

	salt := fullHash[:saltLength]
	storedKey := fullHash[saltLength:]

	// Recompute the hash for the provided password using the same salt
	computedHash := argon2.IDKey([]byte(password), salt, timeCost, memoryCost, parallelism, keyLength)

	// Compare the hashes in constant time to prevent timing attacks
	if subtle.ConstantTimeCompare(storedKey, computedHash) == 1 {
		return nil
	}
	return errors.New("invalid password")
}

// Encrypt encrypts plaintext using AES encryption
func Encrypt(plaintext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)
	return ciphertext, nil
}

// Decrypt decrypts AES-encrypted ciphertext
func Decrypt(ciphertext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)
	return ciphertext, nil
}

// AddUser adds a new user to the credential store
func (cs *CredentialStore) AddUser(username, password, role string) error {
	// Check if the username already exists
	for _, user := range cs.Users {
		if user.Username == username {
			return fmt.Errorf("username %s is already taken", username)
		}
	}

	hashedPassword, err := HashPassword(password)
	if err != nil {
		return err
	}

	user := User{
		Username:       username,
		HashedPassword: hashedPassword,
		Role:           role,
	}

	cs.Users = append(cs.Users, user)
	return nil
}

// UpdateUser updates a user's password in the credential store
func (cs *CredentialStore) UpdateUser(username, newPassword string) error {
	for i, user := range cs.Users {
		if user.Username == username {
			hashedPassword, err := HashPassword(newPassword)
			if err != nil {
				return err
			}
			cs.Users[i].HashedPassword = hashedPassword
			return nil
		}
	}
	return fmt.Errorf("user not found: %s", username)
}

// FindUserByUsername finds a user by username
func (cs *CredentialStore) FindUserByUsername(username string) (*User, error) {
	for _, user := range cs.Users {
		if user.Username == username {
			return &user, nil
		}
	}
	return nil, fmt.Errorf("user not found: %s", username)
}

// SaveCredentials saves the credential store to a file as encrypted JSON
func SaveCredentials(store *CredentialStore, key []byte, filepath string) error {
	data, err := json.Marshal(store)
	if err != nil {
		return err
	}

	encryptedData, err := Encrypt(data, key)
	if err != nil {
		return err
	}

	return os.WriteFile(filepath, encryptedData, 0600)
}

// LoadCredentials loads the credential store from an encrypted JSON file
func LoadCredentials(key []byte, filepath string) (*CredentialStore, error) {
	encryptedData, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	decryptedData, err := Decrypt(encryptedData, key)
	if err != nil {
		return nil, err
	}

	var store CredentialStore
	if err := json.Unmarshal(decryptedData, &store); err != nil {
		return nil, err
	}

	return &store, nil
}

// Package encryption provides secure link generation and management utilities for the song project
package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type EncryptionKey struct {
	Key []byte
}

type MagicLink struct {
	Token      string
	UserID     string
	Expiry     time.Time
	DeviceInfo string
	Used       bool
}

type EncryptionManager struct {
	key *EncryptionKey
}

func NewEncryptionManager(secretKey string) (*EncryptionManager, error) {
	if len(secretKey) < 32 {
		return nil, errors.New("secret key must be at least 32 characters")
	}

	hashedKey := sha256.Sum256([]byte(secretKey))
	return &EncryptionManager{
		key: &EncryptionKey{Key: hashedKey[:32]},
	}, nil
}

func (em *EncryptionManager) GenerateMagicToken() (string, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}

	return base64.URLEncoding.EncodeToString(tokenBytes), nil
}

type EncryptedPayload struct {
	UserID          string
	Expiry          int64
	DeviceSignature string
}

func (em *EncryptionManager) EncryptData(userID string, expiry time.Time, deviceInfo string) (string, error) {
	deviceHash := sha256.Sum256([]byte(deviceInfo))
	deviceSignature := base64.URLEncoding.EncodeToString(deviceHash[:])

	payload := fmt.Sprintf("%s|%d|%s", userID, expiry.Unix(), deviceSignature)

	block, err := aes.NewCipher(em.key.Key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	encrypted := gcm.Seal(nonce, nonce, []byte(payload), nil)
	return base64.URLEncoding.EncodeToString(encrypted), nil
}

func (em *EncryptionManager) DecryptData(encrypted string) (*EncryptedPayload, error) {
	data, err := base64.URLEncoding.DecodeString(encrypted)
	if err != nil {
		return nil, fmt.Errorf("failed to decode encrypted data: %w", err)
	}

	block, err := aes.NewCipher(em.key.Key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	if len(data) < gcm.NonceSize() {
		return nil, errors.New("ciphertext too short")
	}

	nonce := data[:gcm.NonceSize()]
	ciphertext := data[gcm.NonceSize():]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %w", err)
	}

	parts := strings.Split(string(plaintext), "|")
	if len(parts) != 3 {
		return nil, errors.New("invalid encrypted data format")
	}

	expiry, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid expiry timestamp: %w", err)
	}

	return &EncryptedPayload{
		UserID:          parts[0],
		Expiry:          expiry,
		DeviceSignature: parts[2],
	}, nil
}

func (em *EncryptionManager) ValidateLink(link, userID string, deviceInfo string) (*MagicLink, error) {
	if link == "" || userID == "" {
		return nil, errors.New("link and userID are required")
	}

	payload, err := em.DecryptData(link)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt magic link: %w", err)
	}

	if time.Now().Unix() > payload.Expiry {
		return nil, errors.New("magic link has expired")
	}

	if !em.verifyDeviceBinding(payload, deviceInfo) {
		return nil, errors.New("device binding verification failed")
	}

	if payload.UserID != userID {
		return nil, errors.New("user ID mismatch")
	}

	magicLink := &MagicLink{
		Token:      link,
		UserID:     userID,
		Expiry:     time.Unix(payload.Expiry, 0),
		DeviceInfo: deviceInfo,
		Used:       false,
	}

	return magicLink, nil
}

func (em *EncryptionManager) GenerateMagicLink(userID string, deviceInfo string, duration time.Time) (string, error) {
	expiry := time.Now().Add(time.Hour * 24) // Default 24-hour expiry
	if !duration.IsZero() {
		expiry = duration
	}

	encrypted, err := em.EncryptData(userID, expiry, deviceInfo)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt magic link: %w", err)
	}

	return encrypted, nil
}

func (em *EncryptionManager) verifyDeviceBinding(payload *EncryptedPayload, deviceInfo string) bool {
	if deviceInfo == "" {
		return true
	}

	deviceHash := sha256.Sum256([]byte(deviceInfo))
	deviceSignature := base64.URLEncoding.EncodeToString(deviceHash[:])

	return deviceSignature == payload.DeviceSignature
}

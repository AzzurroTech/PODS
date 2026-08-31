// Package auth implements the authentication handlers and services for the song project
package auth

import (
	//"encoding/json"
	"fmt"
	//"log"
	//"net/http"
	"time"

	"azzurrotech/song/internal/encryption"
)

type GenerateLinkRequest struct {
	UserID     string `json:"user_id"`
	DeviceInfo string `json:"device_info,omitempty"`
}

type GenerateLinkResponse struct {
	MagicLink string    `json:"magic_link"`
	ExpiresAt time.Time `json:"expires_at"`
}

type ValidateLinkRequest struct {
	Link       string `json:"link"`
	UserID     string `json:"user_id"`
	DeviceInfo string `json:"device_info,omitempty"`
}

type ValidateLinkResponse struct {
	Valid     bool      `json:"valid"`
	UserID    string    `json:"user_id,omitempty"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
}

type AuthService struct {
	encryptionManager *encryption.EncryptionManager
	issuedLinks       map[string]*encryption.MagicLink
}

func NewAuthService(secretKey string) (*AuthService, error) {
	em, err := encryption.NewEncryptionManager(secretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create encryption manager: %w", err)
	}

	return &AuthService{
		encryptionManager: em,
		issuedLinks:       make(map[string]*encryption.MagicLink),
	}, nil
}

func (s *AuthService) GenerateMagicLink(userID string, deviceInfo string) (*encryption.MagicLink, error) {
	link, err := s.encryptionManager.GenerateMagicLink(userID, deviceInfo, time.Time{})
	if err != nil {
		return nil, fmt.Errorf("failed to generate magic link: %w", err)
	}

	validatedLink, err := s.encryptionManager.ValidateLink(link, userID, deviceInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to validate generated link: %w", err)
	}

	s.issuedLinks[link] = validatedLink

	return validatedLink, nil
}

func (s *AuthService) ValidateMagicLink(link, userID, deviceInfo string) (*encryption.MagicLink, error) {
	cachedLink, exists := s.issuedLinks[link]
	if exists && !cachedLink.Used {
		return cachedLink, nil
	}

	validatedLink, err := s.encryptionManager.ValidateLink(link, userID, deviceInfo)
	if err != nil {
		return nil, err
	}

	s.issuedLinks[link] = validatedLink

	return validatedLink, nil
}

func (s *AuthService) MarkLinkAsUsed(link string) error {
	magicLink, exists := s.issuedLinks[link]
	if !exists {
		return fmt.Errorf("magic link not found")
	}

	if magicLink.Used {
		return fmt.Errorf("magic link already used")
	}

	magicLink.Used = true
	s.issuedLinks[link] = magicLink

	return nil
}

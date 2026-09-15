// Package service contains the application service layer.
//
// Services implement business logic and coordinate between repositories,
// domain models, and other services. They must not import handler packages.
//
// password.go: Password validation and Argon2id hashing.
package service

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/crypto/argon2"
)

// PasswordParams holds the Argon2id parameters used for hashing.
// These are stored alongside the hash to allow future parameter upgrades.
type PasswordParams struct {
	Time    uint32
	Memory  uint32
	Threads uint8
	KeyLen  uint32
}

// DefaultPasswordParams returns OWASP-recommended minimum Argon2id parameters.
// time=2, memory=64MB, threads=4 exceeds the OWASP 2023 minimum.
func DefaultPasswordParams() PasswordParams {
	return PasswordParams{
		Time:    2,
		Memory:  65536, // 64 MiB
		Threads: 4,
		KeyLen:  32,
	}
}

// PasswordService handles password validation and hashing.
type PasswordService struct {
	params PasswordParams
}

// NewPasswordService creates a PasswordService with the given Argon2id parameters.
func NewPasswordService(params PasswordParams) *PasswordService {
	return &PasswordService{params: params}
}

// PasswordStrength categorizes a password's strength.
type PasswordStrength int

const (
	PasswordStrengthWeak   PasswordStrength = iota
	PasswordStrengthFair
	PasswordStrengthGood
	PasswordStrengthStrong
)

// PasswordValidationResult holds the outcome of password validation.
type PasswordValidationResult struct {
	Valid    bool
	Strength PasswordStrength
	Feedback []string // User-facing messages (no technical details)
}

// commonPasswords is a small sample of extremely common passwords.
// A production deployment should use a larger list (e.g., haveibeenpwned API or top-10k list).
var commonPasswords = map[string]bool{
	"password": true, "password1": true, "password123": true,
	"123456": true, "1234567": true, "12345678": true, "123456789": true,
	"qwerty": true, "qwerty123": true, "abc123": true, "letmein": true,
	"welcome": true, "monkey": true, "dragon": true, "master": true,
	"sunshine": true, "princess": true, "shadow": true, "superman": true,
	"iloveyou": true, "trustno1": true, "admin": true, "login": true,
	"passw0rd": true, "p@ssword": true, "p@ssw0rd": true, "secret": true,
}

// repeatingPattern matches strings of a single repeating character (e.g., "aaaaaaa").
var repeatingPattern = regexp.MustCompile(`^(.)\1{5,}$`)

// ValidatePassword checks a password against security policy.
// The result includes user-friendly feedback messages.
// The name and email are checked to prevent obvious personalisation.
func (ps *PasswordService) ValidatePassword(password, name, email string) PasswordValidationResult {
	feedback := []string{}

	// Minimum length: 10 characters.
	// Long passphrases are encouraged — no maximum imposed.
	if len(password) < 10 {
		feedback = append(feedback, "Password must be at least 10 characters long. Consider using a passphrase.")
	}

	// Common password check (case-insensitive).
	if commonPasswords[strings.ToLower(password)] {
		feedback = append(feedback, "This password is too common. Please choose something more unique.")
	}

	// Repeating character check.
	if repeatingPattern.MatchString(password) {
		feedback = append(feedback, "Avoid passwords made of repeated characters.")
	}

	// Sequential character check (e.g., "12345678", "abcdef").
	if isSequential(password) {
		feedback = append(feedback, "Avoid obvious sequences like '12345678' or 'abcdef'.")
	}

	// Personal information check.
	lower := strings.ToLower(password)
	if name != "" && len(name) >= 4 && strings.Contains(lower, strings.ToLower(name)) {
		feedback = append(feedback, "Avoid using your name in your password.")
	}
	if email != "" {
		localPart := strings.Split(email, "@")[0]
		if len(localPart) >= 4 && strings.Contains(lower, strings.ToLower(localPart)) {
			feedback = append(feedback, "Avoid using your email address in your password.")
		}
	}

	strength := assessStrength(password)
	valid := len(feedback) == 0

	// If no blocking issues but strength is weak, add advisory (not blocking).
	if valid && strength == PasswordStrengthFair {
		feedback = append(feedback, "Password is acceptable but could be stronger. Consider a longer passphrase.")
	}

	return PasswordValidationResult{
		Valid:    valid,
		Strength: strength,
		Feedback: feedback,
	}
}

// HashPassword hashes a password using Argon2id with a random salt.
// Returns the encoded hash string in the format:
//   $argon2id$v=19$m=<mem>,t=<time>,p=<threads>$<salt_b64>$<hash_b64>
func (ps *PasswordService) HashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate salt: %w", err)
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		ps.params.Time,
		ps.params.Memory,
		ps.params.Threads,
		ps.params.KeyLen,
	)

	encoded := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		ps.params.Memory,
		ps.params.Time,
		ps.params.Threads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)
	return encoded, nil
}

// VerifyPassword checks whether a plaintext password matches an Argon2id encoded hash.
// Returns false (not an error) if the password is wrong.
func (ps *PasswordService) VerifyPassword(password, encodedHash string) (bool, error) {
	params, salt, hash, err := decodeHash(encodedHash)
	if err != nil {
		return false, fmt.Errorf("decode hash: %w", err)
	}

	candidate := argon2.IDKey(
		[]byte(password),
		salt,
		params.Time,
		params.Memory,
		params.Threads,
		params.KeyLen,
	)

	// Constant-time comparison to prevent timing attacks.
	if subtle.ConstantTimeCompare(hash, candidate) == 1 {
		return true, nil
	}
	return false, nil
}

// decodeHash parses an encoded Argon2id hash string.
func decodeHash(encoded string) (PasswordParams, []byte, []byte, error) {
	parts := strings.Split(encoded, "$")
	// Format: ["", "argon2id", "v=19", "m=...,t=...,p=...", "<salt>", "<hash>"]
	if len(parts) != 6 {
		return PasswordParams{}, nil, nil, fmt.Errorf("invalid hash format: expected 6 parts, got %d", len(parts))
	}
	if parts[1] != "argon2id" {
		return PasswordParams{}, nil, nil, fmt.Errorf("unsupported algorithm: %s", parts[1])
	}

	var params PasswordParams
	_, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &params.Memory, &params.Time, &params.Threads)
	if err != nil {
		return PasswordParams{}, nil, nil, fmt.Errorf("parse argon2id params: %w", err)
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return PasswordParams{}, nil, nil, fmt.Errorf("decode salt: %w", err)
	}

	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return PasswordParams{}, nil, nil, fmt.Errorf("decode hash: %w", err)
	}

	params.KeyLen = uint32(len(hash))
	return params, salt, hash, nil
}

// assessStrength returns a strength classification for a password.
func assessStrength(password string) PasswordStrength {
	length := len(password)
	hasLower := false
	hasUpper := false
	hasDigit := false
	hasSpecial := false

	for _, r := range password {
		switch {
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsDigit(r):
			hasDigit = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r) || unicode.IsSpace(r):
			hasSpecial = true
		}
	}

	variety := 0
	if hasLower {
		variety++
	}
	if hasUpper {
		variety++
	}
	if hasDigit {
		variety++
	}
	if hasSpecial {
		variety++
	}

	switch {
	case length >= 20 && variety >= 3:
		return PasswordStrengthStrong
	case length >= 14 && variety >= 2:
		return PasswordStrengthGood
	case length >= 10:
		return PasswordStrengthFair
	default:
		return PasswordStrengthWeak
	}
}

// isSequential returns true if the password consists largely of sequential characters.
func isSequential(s string) bool {
	if len(s) < 6 {
		return false
	}
	sequential := 0
	for i := 1; i < len(s); i++ {
		if s[i] == s[i-1]+1 || s[i] == s[i-1]-1 {
			sequential++
		}
	}
	return sequential >= len(s)-2
}

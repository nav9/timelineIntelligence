package tests

import (
	"strings"
	"testing"

	"timeline-intelligence/backend/internal/service"
)

// newTestPasswordService creates a PasswordService with fast parameters for tests.
func newTestPasswordService() *service.PasswordService {
	return service.NewPasswordService(service.PasswordParams{
		Time:    1,
		Memory:  8192,
		Threads: 1,
		KeyLen:  32,
	})
}

// --- Hash and verify ---

func TestPassword_HashAndVerify(t *testing.T) {
	svc := newTestPasswordService()
	password := "correct-horse-battery-staple"

	hash, err := svc.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	ok, err := svc.VerifyPassword(password, hash)
	if err != nil {
		t.Fatalf("VerifyPassword error: %v", err)
	}
	if !ok {
		t.Error("VerifyPassword returned false for correct password")
	}
}

func TestPassword_WrongPasswordFails(t *testing.T) {
	svc := newTestPasswordService()

	hash, err := svc.HashPassword("correct-horse-battery-staple")
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	ok, err := svc.VerifyPassword("wrong-password-here!", hash)
	if err != nil {
		t.Fatalf("VerifyPassword error: %v", err)
	}
	if ok {
		t.Error("VerifyPassword returned true for wrong password")
	}
}

func TestPassword_HashIsArgon2id(t *testing.T) {
	svc := newTestPasswordService()

	hash, err := svc.HashPassword("correct-horse-battery-staple")
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if !strings.HasPrefix(hash, "$argon2id$") {
		t.Errorf("Expected Argon2id hash prefix, got: %s", hash[:20])
	}
}

func TestPassword_HashesAreDifferentEachTime(t *testing.T) {
	svc := newTestPasswordService()
	password := "correct-horse-battery-staple"

	hash1, _ := svc.HashPassword(password)
	hash2, _ := svc.HashPassword(password)

	if hash1 == hash2 {
		t.Error("Two hashes of the same password must differ (different random salts)")
	}
}

func TestPassword_NeverStoredInHash(t *testing.T) {
	svc := newTestPasswordService()
	password := "correct-horse-battery-staple"

	hash, _ := svc.HashPassword(password)
	if strings.Contains(hash, password) {
		t.Error("Password appears in plaintext within its own hash")
	}
}

// --- Validation ---

func TestPasswordValidation_ShortPassword(t *testing.T) {
	svc := newTestPasswordService()
	result := svc.ValidatePassword("short", "", "")
	if result.Valid {
		t.Error("Expected short password to be invalid")
	}
}

func TestPasswordValidation_CommonPassword(t *testing.T) {
	svc := newTestPasswordService()
	result := svc.ValidatePassword("password123", "", "")
	if result.Valid {
		t.Error("Expected common password to be invalid")
	}
}

func TestPasswordValidation_RepeatingChars(t *testing.T) {
	svc := newTestPasswordService()
	result := svc.ValidatePassword("aaaaaaaaaa", "", "")
	if result.Valid {
		t.Error("Expected repeating-character password to be invalid")
	}
}

func TestPasswordValidation_SequentialChars(t *testing.T) {
	svc := newTestPasswordService()
	result := svc.ValidatePassword("1234567890", "", "")
	if result.Valid {
		t.Error("Expected sequential password to be invalid")
	}
}

func TestPasswordValidation_ContainsName(t *testing.T) {
	svc := newTestPasswordService()
	result := svc.ValidatePassword("alice-is-great-2024", "alice", "other@example.com")
	if result.Valid {
		t.Error("Expected password containing user's name to be invalid")
	}
}

func TestPasswordValidation_ContainsEmailLocal(t *testing.T) {
	svc := newTestPasswordService()
	result := svc.ValidatePassword("testuser-is-great-2024", "Other Name", "testuser@example.com")
	if result.Valid {
		t.Error("Expected password containing email local part to be invalid")
	}
}

func TestPasswordValidation_StrongPassphrase(t *testing.T) {
	svc := newTestPasswordService()
	result := svc.ValidatePassword("correct-horse-battery-staple!", "Alice", "alice@example.com")
	if !result.Valid {
		t.Errorf("Expected strong passphrase to be valid; feedback: %v", result.Feedback)
	}
}

func TestPasswordValidation_LongPassphraseStrong(t *testing.T) {
	svc := newTestPasswordService()
	result := svc.ValidatePassword("my-very-secure-passphrase-2024!", "Bob", "bob@example.com")
	if result.Strength < service.PasswordStrengthGood {
		t.Errorf("Expected strong/good strength for long passphrase, got %d", result.Strength)
	}
}

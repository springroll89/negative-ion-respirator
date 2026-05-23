package service_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"negative-ion-respirator/backend/internal/model"
	"negative-ion-respirator/backend/internal/service"
)

// mockAdminRepo implements service.AdminRepo for testing.
type mockAdminRepo struct {
	users map[string]*model.AdminUser
}

func (m *mockAdminRepo) FindByUsername(_ context.Context, username string) (*model.AdminUser, error) {
	if u, ok := m.users[username]; ok {
		return u, nil
	}
	return nil, sql.ErrNoRows
}

func (m *mockAdminRepo) Create(_ context.Context, u *model.AdminUser) error {
	m.users[u.Username] = u
	return nil
}

// hashPassword is a test helper for generating bcrypt hashes.
func hashPassword(t *testing.T, password string) string {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("bcrypt hash: %v", err)
	}
	return string(hash)
}

func TestAuthService_Login_Success(t *testing.T) {
	const testSecret = "test-jwt-secret-key-for-tests"
	password := "secure-password"

	repo := &mockAdminRepo{
		users: map[string]*model.AdminUser{
			"admin": {
				ID:           1,
				Username:     "admin",
				PasswordHash: hashPassword(t, password),
				Role:         "admin",
				CreatedAt:    time.Now(),
			},
		},
	}
	svc := service.NewAuthService(repo, testSecret)

	resp, err := svc.Login(context.Background(), model.LoginReq{
		Username: "admin",
		Password: password,
	})
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	if resp.Token == "" {
		t.Error("expected non-empty token")
	}
	if resp.ExpiresAt == 0 {
		t.Error("expected non-zero expires_at")
	}

	// Verify the token is valid
	claims, err := svc.ValidateToken(resp.Token)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}
	if claims["username"] != "admin" {
		t.Errorf("expected username 'admin', got '%v'", claims["username"])
	}
	if claims["role"] != "admin" {
		t.Errorf("expected role 'admin', got '%v'", claims["role"])
	}
}

func TestAuthService_Login_InvalidPassword(t *testing.T) {
	const testSecret = "test-jwt-secret-key-for-tests"

	repo := &mockAdminRepo{
		users: map[string]*model.AdminUser{
			"admin": {
				ID:           1,
				Username:     "admin",
				PasswordHash: hashPassword(t, "correct-password"),
				Role:         "admin",
				CreatedAt:    time.Now(),
			},
		},
	}
	svc := service.NewAuthService(repo, testSecret)

	_, err := svc.Login(context.Background(), model.LoginReq{
		Username: "admin",
		Password: "wrong-password",
	})
	if err == nil {
		t.Error("expected error for invalid password")
	}
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	const testSecret = "test-jwt-secret-key-for-tests"

	repo := &mockAdminRepo{
		users: make(map[string]*model.AdminUser),
	}
	svc := service.NewAuthService(repo, testSecret)

	_, err := svc.Login(context.Background(), model.LoginReq{
		Username: "nonexistent",
		Password: "anything",
	})
	if err == nil {
		t.Error("expected error for non-existent user")
	}
}

func TestAuthService_ValidateToken_InvalidToken(t *testing.T) {
	const testSecret = "test-jwt-secret-key-for-tests"

	repo := &mockAdminRepo{users: make(map[string]*model.AdminUser)}
	svc := service.NewAuthService(repo, testSecret)

	_, err := svc.ValidateToken("not-a-valid-jwt")
	if err == nil {
		t.Error("expected error for invalid token")
	}
}

func TestAuthService_ValidateToken_WrongSecret(t *testing.T) {
	repo := &mockAdminRepo{users: make(map[string]*model.AdminUser)}
	svc1 := service.NewAuthService(repo, "secret-one")
	svc2 := service.NewAuthService(repo, "secret-two")

	// Create a token with svc1's secret
	resp, err := svc1.Login(context.Background(), model.LoginReq{
		Username: "admin",
		Password: hashPassword(t, "pw"),
	})
	if err != nil {
		// User won't exist, so generate token manually
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id":  1,
			"username": "admin",
			"role":     "admin",
			"exp":      time.Now().Add(24 * time.Hour).Unix(),
		})
		tokenStr, _ := token.SignedString([]byte("secret-one"))

		_, err = svc2.ValidateToken(tokenStr)
		if err == nil {
			t.Error("expected error when validating token with wrong secret")
		}
		return
	}
	_ = resp

	// Fallback test with manual token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  1,
		"username": "admin",
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	})
	tokenStr, _ := token.SignedString([]byte("secret-one"))

	_, err2 := svc2.ValidateToken(tokenStr)
	if err2 == nil {
		t.Error("expected error when validating token with wrong secret")
	}
}

func TestAuthService_RefreshToken_Success(t *testing.T) {
	const testSecret = "test-jwt-secret-key-for-tests"

	repo := &mockAdminRepo{
		users: map[string]*model.AdminUser{
			"admin": {
				ID:       1,
				Username: "admin",
				Role:     "admin",
			},
		},
	}
	svc := service.NewAuthService(repo, testSecret)

	resp, err := svc.RefreshToken(context.Background(), "admin")
	if err != nil {
		t.Fatalf("RefreshToken failed: %v", err)
	}
	if resp.Token == "" {
		t.Error("expected non-empty refreshed token")
	}
}

func TestAuthService_RefreshToken_UserNotFound(t *testing.T) {
	const testSecret = "test-jwt-secret-key-for-tests"

	repo := &mockAdminRepo{
		users: make(map[string]*model.AdminUser),
	}
	svc := service.NewAuthService(repo, testSecret)

	_, err := svc.RefreshToken(context.Background(), "nonexistent")
	if err == nil {
		t.Error("expected error for non-existent user")
	}
}

func TestAuthService_ValidateToken_ExpiredToken(t *testing.T) {
	const testSecret = "test-jwt-secret-key-for-tests"

	repo := &mockAdminRepo{users: make(map[string]*model.AdminUser)}
	svc := service.NewAuthService(repo, testSecret)

	// Create a token that expired 1 hour ago
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  1,
		"username": "admin",
		"role":     "admin",
		"exp":      time.Now().Add(-1 * time.Hour).Unix(),
	})
	tokenStr, err := token.SignedString([]byte(testSecret))
	if err != nil {
		t.Fatalf("failed to create expired token: %v", err)
	}

	_, err = svc.ValidateToken(tokenStr)
	if err == nil {
		t.Error("expected error for expired token")
	} else {
		t.Logf("expected error: %v", err)
	}
}

func TestAuthService_TokenFlow(t *testing.T) {
	// Integration-style test that validates the complete auth flow:
	// login -> validate -> refresh -> validate new token
	const testSecret = "test-jwt-secret"

	password := "flow-password"
	repo := &mockAdminRepo{
		users: map[string]*model.AdminUser{
			"admin": {
				ID:           1,
				Username:     "admin",
				PasswordHash: hashPassword(t, password),
				Role:         "admin",
				CreatedAt:    time.Now(),
			},
		},
	}
	svc := service.NewAuthService(repo, testSecret)

	// Step 1: Login
	loginResp, err := svc.Login(context.Background(), model.LoginReq{
		Username: "admin", Password: password,
	})
	if err != nil {
		t.Fatalf("login step failed: %v", err)
	}

	// Step 2: Validate the login token
	claims, err := svc.ValidateToken(loginResp.Token)
	if err != nil {
		t.Fatalf("validate after login failed: %v", err)
	}
	if claims["username"] != "admin" {
		t.Errorf("expected 'admin', got '%v'", claims["username"])
	}

	// Step 3: Refresh token
	refreshResp, err := svc.RefreshToken(context.Background(), "admin")
	if err != nil {
		t.Fatalf("refresh step failed: %v", err)
	}

	// Step 4: Validate the refreshed token
	claims2, err := svc.ValidateToken(refreshResp.Token)
	if err != nil {
		t.Fatalf("validate after refresh failed: %v", err)
	}
	if claims2["username"] != "admin" {
		t.Errorf("expected 'admin' in refreshed token, got '%v'", claims2["username"])
	}

	// Note: JWT tokens may be identical when issued within the same second
	// with identical claims. In production, add a jti (JWT ID) claim with a
	// random value to ensure uniqueness.
	if loginResp.Token == refreshResp.Token {
		t.Log("note: login and refresh tokens are identical (same-second issue); " +
			"a jti claim would ensure uniqueness in production")
	}

	t.Logf("%s", fmt.Sprintf("token flow: login=%t, validate=%t, refresh=%t",
		loginResp.Token != "",
		claims["username"] == "admin",
		refreshResp.Token != ""))
}

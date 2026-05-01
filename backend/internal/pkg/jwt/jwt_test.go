package jwt

import (
	"akatengu/internal/model/db"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func newTestUser() *db.User {
	return &db.User{
		UserId:   42,
		Username: "testuser",
	}
}

func TestGenerateToken_ReturnsNonEmptyString(t *testing.T) {
	svc := NewJWT("secret")
	token, err := svc.GenerateToken(newTestUser())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}
}

func TestGenerateToken_IsValidJWTFormat(t *testing.T) {
	svc := NewJWT("secret")
	token, err := svc.GenerateToken(newTestUser())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("expected 3 JWT parts, got %d", len(parts))
	}
}

func TestVerifyToken_RoundTrip(t *testing.T) {
	svc := NewJWT("secret")
	user := newTestUser()
	token, err := svc.GenerateToken(user)
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}

	claims, err := svc.VerifyToken(token)
	if err != nil {
		t.Fatalf("VerifyToken: %v", err)
	}
	if claims.UserID != user.UserId {
		t.Errorf("UserID: want %d, got %d", user.UserId, claims.UserID)
	}
	if claims.UserName != user.Username {
		t.Errorf("UserName: want %q, got %q", user.Username, claims.UserName)
	}
}

func TestVerifyToken_WrongKey(t *testing.T) {
	svc := NewJWT("secret")
	token, err := svc.GenerateToken(newTestUser())
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}

	other := NewJWT("wrong-secret")
	_, err = other.VerifyToken(token)
	if err == nil {
		t.Fatal("expected error for token signed with different key")
	}
}

func TestVerifyToken_TamperedToken(t *testing.T) {
	svc := NewJWT("secret")
	token, err := svc.GenerateToken(newTestUser())
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}

	tampered := token + "x"
	_, err = svc.VerifyToken(tampered)
	if err == nil {
		t.Fatal("expected error for tampered token")
	}
}

func TestVerifyToken_EmptyString(t *testing.T) {
	svc := NewJWT("secret")
	_, err := svc.VerifyToken("")
	if err == nil {
		t.Fatal("expected error for empty token")
	}
}

func TestVerifyToken_MalformedToken(t *testing.T) {
	svc := NewJWT("secret")
	_, err := svc.VerifyToken("not.a.jwt")
	if err == nil {
		t.Fatal("expected error for malformed token")
	}
}

func TestVerifyToken_ExpiredToken(t *testing.T) {
	svc := &jwtService{jwtKey: []byte("secret")}
	user := newTestUser()

	claims := Claims{
		UserID:   user.UserId,
		UserName: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Minute)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(svc.jwtKey)
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	_, err = svc.VerifyToken(signed)
	if err == nil {
		t.Fatal("expected error for expired token")
	}
}

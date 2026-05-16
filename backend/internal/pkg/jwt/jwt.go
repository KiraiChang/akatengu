package jwt

import (
	"akatengu/internal/model/db"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JwtService interface {
	GenerateToken(user *db.User) (string, error)
	GenerateTokenWithMerchant(user *db.User, merchantID int64, role string) (string, error)
	VerifyToken(tokenStr string) (*Claims, error)
}

type jwtService struct {
	jwtKey []byte
}

type Claims struct {
	UserID     int64  `json:"uid"`
	UserName   string `json:"user_name"`
	MerchantID int64  `json:"mid"`
	Role       string `json:"role"`
	jwt.RegisteredClaims
}

func NewJWT(jwtKey string) JwtService {
	return &jwtService{
		jwtKey: []byte(jwtKey),
	}
}

func (j *jwtService) GenerateToken(user *db.User) (string, error) {
	claims := Claims{
		UserID:   user.UserId,
		UserName: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * time.Minute)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.jwtKey)
}

func (j *jwtService) GenerateTokenWithMerchant(user *db.User, merchantID int64, role string) (string, error) {
	claims := Claims{
		UserID:     user.UserId,
		UserName:   user.Username,
		MerchantID: merchantID,
		Role:       role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * time.Minute)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.jwtKey)
}

func (j *jwtService) VerifyToken(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return j.jwtKey, nil
	})
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("unauthorized")
	}
	return claims, nil
}

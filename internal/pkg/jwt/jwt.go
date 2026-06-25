package jwt

import (
	"errors"
	"time"

	"appliance-recycle/internal/config"

	"github.com/golang-jwt/jwt/v4"
)

type UserType string

const (
	UserTypeResident UserType = "resident"
	UserTypeAdmin    UserType = "admin"
)

type Claims struct {
	UserID   uint64   `json:"user_id"`
	UserType UserType `json:"user_type"`
	jwt.RegisteredClaims
}

func GenerateResidentToken(userID uint64) (string, error) {
	return generateToken(userID, UserTypeResident, config.Cfg.JWT.ResidentSecret, config.Cfg.JWT.ExpireHours)
}

func GenerateAdminToken(userID uint64) (string, error) {
	return generateToken(userID, UserTypeAdmin, config.Cfg.JWT.AdminSecret, config.Cfg.JWT.ExpireHours)
}

func generateToken(userID uint64, userType UserType, secret string, expireHours int) (string, error) {
	claims := Claims{
		UserID:   userID,
		UserType: userType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expireHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "appliance-recycle",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func ParseResidentToken(tokenString string) (*Claims, error) {
	return parseToken(tokenString, config.Cfg.JWT.ResidentSecret, UserTypeResident)
}

func ParseAdminToken(tokenString string) (*Claims, error) {
	return parseToken(tokenString, config.Cfg.JWT.AdminSecret, UserTypeAdmin)
}

func parseToken(tokenString string, secret string, expectedType UserType) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})

	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				return nil, errors.New("malformed")
			}
			if ve.Errors&jwt.ValidationErrorExpired != 0 {
				return nil, errors.New("expired")
			}
			if ve.Errors&jwt.ValidationErrorNotValidYet != 0 {
				return nil, errors.New("not valid yet")
			}
		}
		return nil, errors.New("invalid")
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		if claims.UserType != expectedType {
			return nil, errors.New("wrong user type")
		}
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

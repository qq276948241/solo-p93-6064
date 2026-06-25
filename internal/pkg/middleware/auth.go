package middleware

import (
	"net/http"
	"strings"

	"appliance-recycle/internal/pkg/jwt"
	"appliance-recycle/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

type contextKey string

const (
	ContextUserID   contextKey = "user_id"
	ContextUserType contextKey = "user_type"
)

func ResidentAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, errCode := parseToken(c, jwt.UserTypeResident)
		if errCode != 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, response.Fail(errCode))
			return
		}
		c.Set(string(ContextUserID), claims.UserID)
		c.Set(string(ContextUserType), claims.UserType)
		c.Next()
	}
}

func AdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, errCode := parseToken(c, jwt.UserTypeAdmin)
		if errCode != 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, response.Fail(errCode))
			return
		}
		c.Set(string(ContextUserID), claims.UserID)
		c.Set(string(ContextUserType), claims.UserType)
		c.Next()
	}
}

func parseToken(c *gin.Context, userType jwt.UserType) (*jwt.Claims, int) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return nil, response.CodeTokenInvalid
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if !(len(parts) == 2 && parts[0] == "Bearer") {
		return nil, response.CodeTokenInvalid
	}

	tokenStr := parts[1]
	var claims *jwt.Claims
	var err error

	switch userType {
	case jwt.UserTypeResident:
		claims, err = jwt.ParseResidentToken(tokenStr)
	case jwt.UserTypeAdmin:
		claims, err = jwt.ParseAdminToken(tokenStr)
	default:
		return nil, response.CodeTokenInvalid
	}

	if err != nil {
		switch err.Error() {
		case "expired":
			return nil, response.CodeTokenExpired
		case "malformed":
			return nil, response.CodeTokenMalformed
		default:
			return nil, response.CodeTokenInvalid
		}
	}

	return claims, 0
}

func GetUserID(c *gin.Context) uint64 {
	v, _ := c.Get(string(ContextUserID))
	if id, ok := v.(uint64); ok {
		return id
	}
	return 0
}

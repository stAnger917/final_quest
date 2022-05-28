package authMW

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"time"
)

type Session struct {
	UserID int
	Expiry time.Time
}

var Sessions = map[string]Session{}

func (s Session) IsExpired() bool {
	return s.Expiry.Before(time.Now())
}

func CreateToken(userId int) (string, time.Time) {
	sessionToken := uuid.NewString()
	expiresAt := time.Now().Add(6 * time.Hour)
	Sessions[sessionToken] = Session{
		UserID: userId,
		Expiry: expiresAt,
	}
	return sessionToken, expiresAt
}

func GetLoginFromToken(sessionToken string) int {
	userSession, _ := Sessions[sessionToken]
	return userSession.UserID
}

func CheckToken(sessionToken string) (bool, error) {
	userSession, exists := Sessions[sessionToken]
	if !exists {
		return false, nil
	}
	if userSession.IsExpired() {
		return false, errors.New("token is expired")
	}
	return true, nil
}

func TokenMW() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie("session_token")
		if err != nil {
			c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token!"})
			c.Abort()
			return
		}
		tokenStatus, err := CheckToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, map[string]string{"error": "token is expired!"})
			c.Abort()
			return
		}
		if !tokenStatus {
			delete(Sessions, token)
			c.JSON(http.StatusUnauthorized, map[string]string{"error": "user`s session not found!"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func PrepareTestTokens() {
	Sessions["test_token1"] = Session{
		UserID: 1,
		Expiry: time.Now().Add(24 * time.Hour),
	}
	Sessions["test_token123"] = Session{
		UserID: 2,
		Expiry: time.Now().Add(24 * time.Hour),
	}
	Sessions["test_token1234"] = Session{
		UserID: 2234,
		Expiry: time.Now().Add(24 * time.Hour),
	}
}

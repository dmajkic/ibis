package ibis

import (
	"fmt"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

// Default token lifetime set to 5 days and 56 seconds
const TOKEN_LIFETIME = time.Hour*24*5 + time.Second*56

// JWT token generation
func (s *Server) GenerateToken(user_id interface{}, system interface{}) (string, *time.Time, error) {

	if user_id == "" {
		return "", nil, fmt.Errorf("User unknown.")
	}

	token := jwt.New(jwt.SigningMethodHS256)
	exp := time.Now().Add(TOKEN_LIFETIME)

	// Set some claims
	token.Claims["ID"] = user_id
	token.Claims["sys"] = system
	token.Claims["exp"] = exp.Unix()

	// Sign and get the complete encoded token as a string
	tokenString, err := token.SignedString([]byte(s.authToken))
	return tokenString, &exp, err
}

// Validate token form http request
func (s *Server) CheckToken(request *http.Request) (*jwt.Token, error) {
	token, err := jwt.ParseFromRequest(request, func(token *jwt.Token) (interface{}, error) {

		if token, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Alg)
		}

		return []byte(s.authToken), nil
	})

	return token, err
}

// Authenticate using JWT tokens
func (s *Server) AuthJWT(secret string) gin.HandlerFunc {
	s.authToken = secret
	return func(c *gin.Context) {
		token, err := s.CheckToken(c.Request)
		if err != nil {
			c.AbortWithError(401, err)
			return
		}

		c.Set("user_id", token.Claims["ID"])
	}
}

// Login handler to login user. Actual login is performed in user app,
// and it is expectd to return valid "id" field. If login went Ok,
// then valid JWT token is generated.
func (s *Server) JWTloginHandler(c *gin.Context) {

	user := make(map[string]interface{})

	if err := s.App.LoginUser(c, user); err != nil {
		JSONError(c, 401, fmt.Errorf("Auth failed"))
		return
	}

	if _, ok := user["id"]; !ok {
		JSONError(c, 401, fmt.Errorf("Auth failed"))
		return
	}

	// Sign and get the complete encoded token as a string
	tokenString, exp, err := s.GenerateToken(user["id"], user["sys"])
	if err != nil {
		JSONError500(c, fmt.Errorf("Could not generate token: %v", err))
		return
	}

	attrs := gin.H{"expires_at": exp}
	for k, v := range user {
		attrs[k] = v
	}

	c.JSON(200, gin.H{
		"id":         tokenString,
		"type":       "token",
		"attributes": attrs,
	})
}

// Valid token can be renewed
func (s *Server) JWTrenewHandler(c *gin.Context) {

	token, err := s.CheckToken(c.Request)

	tokenString, exp, err := s.GenerateToken(token.Claims["ID"].(string), token.Claims["sys"].(bool))
	if err != nil {
		JSONError500(c, fmt.Errorf("Could not generate token"))
		return
	}

	c.JSON(200, gin.H{
		"id":   tokenString,
		"type": "token",
		"attributes": gin.H{
			"expires_at": exp,
		},
	})
}

// logoutHandler invalidates JWT token
// TODO: Invalidate JWT token
func (l *Server) JWTlogoutHandler(c *gin.Context) {
	c.AbortWithStatus(http.StatusNoContent)
}

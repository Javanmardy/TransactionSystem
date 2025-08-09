package auth

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	jwtSecret = []byte(getenv("JWT_SECRET", "CHANGE_ME_IN_PROD"))
	jwtIss    = getenv("JWT_ISS", "transaction-system")
	jwtAud    = getenv("JWT_AUD", "transaction-system-clients")
	jwtTTL    = durMinutes(getenv("JWT_TTL_MINUTES", "60"))
)

var JwtSecret = jwtSecret

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func durMinutes(s string) time.Duration {
	d, err := time.ParseDuration(s + "m")
	if err != nil {
		return 60 * time.Minute
	}
	return d
}

func GenerateJWT(userID int, username, role string) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"user_id":  userID,
		"username": username,
		"role":     role,
		"iss":      jwtIss,
		"aud":      jwtAud,
		"iat":      now.Unix(),
		"nbf":      now.Unix(),
		"exp":      now.Add(jwtTTL).Unix(),
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString(jwtSecret)
}

func ParseJWT(tokenString string) (jwt.MapClaims, error) {
	parser := jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))

	token, err := parser.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}
	if claims["iss"] != jwtIss {
		return nil, errors.New("invalid issuer")
	}
	if aud, ok := claims["aud"].(string); !ok || aud != jwtAud {
		return nil, errors.New("invalid audience")
	}
	return claims, nil
}

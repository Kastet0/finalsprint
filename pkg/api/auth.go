package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
)

type SignInRequest struct {
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token"`
}

func getJWTSecret(password string) []byte {
	return []byte(password)
}

func passwordHash(password string) string {
	hasher := sha256.New()
	hasher.Write([]byte(password))
	return hex.EncodeToString(hasher.Sum(nil))
}

func createToken(password string) (string, error) {
	hash := passwordHash(password)

	expirationTime := time.Now().Add(7 * 24 * time.Hour)

	claims := jwt.MapClaims{
		"passhash": hash,
		"exp":      expirationTime.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(getJWTSecret(password))
	if err != nil {
		return "", fmt.Errorf("error signing token: %w", err)
	}

	return tokenString, nil
}

func validateToken(tokenString, password string) (bool, error) {
	if tokenString == "" {
		return false, errors.New("token is empty")
	}

	expectedHash := passwordHash(password)

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return getJWTSecret(password), nil
	})

	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorExpired != 0 {
				return false, errors.New("token expired")
			}
		}
		return false, fmt.Errorf("token parse error: %w", err)
	}

	if !token.Valid {
		return false, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return false, errors.New("invalid claims type")
	}

	hashClaim, ok := claims["passhash"].(string)
	if !ok || hashClaim != expectedHash {
		return false, errors.New("password hash mismatch (token obsolete or tampered)")
	}

	return true, nil
}

func setAuthCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
		HttpOnly: true,
		Secure:   false,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})
}

func auth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		envPass := os.Getenv("TODO_PASSWORD")

		if len(envPass) == 0 {
			next(w, r)
			return
		}

		cookie, err := r.Cookie("token")
		var jwtToken string
		if err == nil {
			jwtToken = cookie.Value
		}

		if len(jwtToken) > 0 {
			valid, err := validateToken(jwtToken, envPass)

			if !valid {
				if err != nil {
					fmt.Printf("Auth failed: %v\n", err)
				}
				http.Error(w, "Authentification required", http.StatusUnauthorized)
				return
			}
		} else {
			http.Error(w, "Authentification required", http.StatusUnauthorized)
			return
		}

		next(w, r)
	})
}

func signInHandler(w http.ResponseWriter, r *http.Request) {
	envPass := os.Getenv("TODO_PASSWORD")

	if r.Method != http.MethodPost {
		writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if len(envPass) == 0 {
		writeJSON(w, AuthResponse{Token: ""}, http.StatusOK)
		return
	}

	var req SignInRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "JSON deserialization error. Make sure you sent a POST request with Content-Type: application/json and a body like {\"password\": \"YOUR_PASSWORD\"}.", http.StatusBadRequest)
		return
	}

	if req.Password != envPass {
		writeError(w, "Invalid password", http.StatusUnauthorized)
		return
	}

	token, err := createToken(envPass)
	if err != nil {
		fmt.Printf("Error creating token: %v\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	setAuthCookie(w, token)
	writeJSON(w, AuthResponse{Token: token}, http.StatusOK)
}

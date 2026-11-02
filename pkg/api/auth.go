package api

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	tokenCookie = "token"
	// Совпадает со временем жизни куки, которую ставит фронтенд.
	tokenTTL = 8 * time.Hour
)

// password возвращает пароль из TODO_PASSWORD. Пустой пароль
// означает, что аутентификация выключена.
func password() string {
	return os.Getenv("TODO_PASSWORD")
}

// passwordHash — контрольная сумма пароля, которая кладётся в токен.
// Если пароль поменяется, токены со старой суммой станут невалидными.
func passwordHash(pass string) string {
	sum := sha256.Sum256([]byte(pass))
	return hex.EncodeToString(sum[:])
}

// signingKey выводит ключ подписи из пароля, поэтому смена пароля
// заодно делает недействительной подпись всех выданных токенов.
func signingKey(pass string) []byte {
	sum := sha256.Sum256([]byte("todo-jwt-signing-key:" + pass))
	return sum[:]
}

type tokenClaims struct {
	PasswordHash string `json:"pwd"`
	jwt.RegisteredClaims
}

func makeToken(pass string, now time.Time) (string, error) {
	claims := tokenClaims{
		PasswordHash: passwordHash(pass),
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(tokenTTL)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(signingKey(pass))
}

// validToken проверяет подпись, срок действия и соответствие токена паролю pass.
func validToken(token, pass string) bool {
	var claims tokenClaims
	_, err := jwt.ParseWithClaims(token, &claims, func(*jwt.Token) (any, error) {
		return signingKey(pass), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithExpirationRequired())
	if err != nil {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(claims.PasswordHash), []byte(passwordHash(pass))) == 1
}

func signinHandler(w http.ResponseWriter, r *http.Request) {
	pass := password()
	if pass == "" {
		writeError(w, http.StatusBadRequest, "аутентификация не настроена")
		return
	}

	var req struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "некорректный JSON: "+err.Error())
		return
	}
	if subtle.ConstantTimeCompare([]byte(req.Password), []byte(pass)) != 1 {
		writeError(w, http.StatusUnauthorized, "неверный пароль")
		return
	}

	token, err := makeToken(pass, time.Now())
	if err != nil {
		log.Printf("формирование токена: %v", err)
		writeError(w, http.StatusInternalServerError, "не удалось сформировать токен")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}

// auth пропускает запрос дальше, только если пароль не задан
// или в куке token лежит валидный для текущего пароля токен.
func auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if pass := password(); pass != "" {
			var token string
			if cookie, err := r.Cookie(tokenCookie); err == nil {
				token = cookie.Value
			}
			if !validToken(token, pass) {
				writeError(w, http.StatusUnauthorized, "требуется аутентификация")
				return
			}
		}
		next(w, r)
	}
}

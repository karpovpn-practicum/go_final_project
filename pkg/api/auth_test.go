package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestValidToken(t *testing.T) {
	now := time.Now()
	token, err := makeToken("secret", now)
	if err != nil {
		t.Fatal(err)
	}
	expired, err := makeToken("secret", now.Add(-tokenTTL-time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	// Токен с верной суммой пароля, но без подписи.
	unsigned, err := jwt.NewWithClaims(jwt.SigningMethodNone, tokenClaims{
		PasswordHash:     passwordHash("secret"),
		RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour))},
	}).SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name, token, pass string
		want              bool
	}{
		{"верный токен", token, "secret", true},
		{"пароль сменился", token, "other", false},
		{"истёк срок", expired, "secret", false},
		{"без подписи", unsigned, "secret", false},
		{"испорченная подпись", token[:len(token)-2] + "xx", "secret", false},
		{"пустой токен", "", "secret", false},
	}
	for _, tt := range tests {
		if got := validToken(tt.token, tt.pass); got != tt.want {
			t.Errorf("%s: validToken = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestAuthMiddleware(t *testing.T) {
	handler := auth(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})
	call := func(cookie string) int {
		req := httptest.NewRequest(http.MethodGet, "/api/tasks", nil)
		if cookie != "" {
			req.AddCookie(&http.Cookie{Name: tokenCookie, Value: cookie})
		}
		rec := httptest.NewRecorder()
		handler(rec, req)
		return rec.Code
	}

	t.Setenv("TODO_PASSWORD", "")
	if code := call(""); code != http.StatusTeapot {
		t.Errorf("без пароля: код %d, запрос должен пройти", code)
	}

	t.Setenv("TODO_PASSWORD", "secret")
	if code := call(""); code != http.StatusUnauthorized {
		t.Errorf("без куки: код %d, want 401", code)
	}
	token, _ := makeToken("secret", time.Now())
	if code := call(token); code != http.StatusTeapot {
		t.Errorf("с валидным токеном: код %d, запрос должен пройти", code)
	}
}

func TestSignin(t *testing.T) {
	t.Setenv("TODO_PASSWORD", "secret")
	call := func(body string) (int, string) {
		req := httptest.NewRequest(http.MethodPost, "/api/signin", strings.NewReader(body))
		rec := httptest.NewRecorder()
		signinHandler(rec, req)
		return rec.Code, rec.Body.String()
	}

	if code, body := call(`{"password":"secret"}`); code != http.StatusOK || !strings.Contains(body, `"token"`) {
		t.Errorf("верный пароль: %d %s", code, body)
	}
	if code, body := call(`{"password":"wrong"}`); code != http.StatusUnauthorized || !strings.Contains(body, `"error"`) {
		t.Errorf("неверный пароль: %d %s", code, body)
	}
	if code, _ := call(`{bad`); code != http.StatusBadRequest {
		t.Errorf("битый JSON: код %d", code)
	}
}

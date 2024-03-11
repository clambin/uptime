package auth_test

import (
	"github.com/clambin/uptime/pkg/auth"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthenticate(t *testing.T) {
	const validKey = "1234"
	tests := []struct {
		name     string
		token    string
		wantCode int
	}{
		{
			name:     "valid",
			token:    validKey,
			wantCode: http.StatusOK,
		},
		{
			name:     "missing",
			wantCode: http.StatusUnauthorized,
		},
		{
			name:     "invalid",
			token:    validKey + "5678",
			wantCode: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r, _ := http.NewRequest(http.MethodGet, "", nil)
			if tt.token != "" {
				r.Header.Set("Authorization", "Bearer "+tt.token)
			}
			w := httptest.NewRecorder()

			auth.Authenticate(validKey)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			})).ServeHTTP(w, r)

			assert.Equal(t, tt.wantCode, w.Code)
		})
	}
}

package handlers

import (
	"bytes"
	"github.com/clambin/go-common/set"
	"github.com/clambin/uptime/pkg/logtester"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"net/http"
	"net/url"
	"testing"
	"time"
)

func Test_parseRequest(t *testing.T) {
	tests := []struct {
		name     string
		rawQuery string
		wantErr  assert.ErrorAssertionFunc
		wantReq  Request
	}{
		{
			name:    "empty",
			wantErr: assert.Error,
		},
		{
			name:     "invalid",
			rawQuery: `;`,
			wantErr:  assert.Error,
		},
		{
			name:     "valid",
			rawQuery: `target=http://localhost:8080/metrics&method=HEAD&codes=200,403&interval=1m`,
			wantErr:  assert.NoError,
			wantReq: Request{
				Target:     "http://localhost:8080/metrics",
				Method:     http.MethodHead,
				ValidCodes: set.New(http.StatusOK, http.StatusForbidden),
				Interval:   1 * time.Minute,
			},
		},
		{
			name:     "target is mandatory",
			rawQuery: `method=GET&codes=200&interval=5m`,
			wantErr:  assert.Error,
		},
		{
			name:     "defaults",
			rawQuery: `target=http://localhost:8080/metrics`,
			wantErr:  assert.NoError,
			wantReq: Request{
				Target:     "http://localhost:8080/metrics",
				Method:     http.MethodGet,
				ValidCodes: set.New(http.StatusOK),
				Interval:   5 * time.Minute,
			},
		},
		{
			name:     "invalid code",
			rawQuery: `target=http://localhost:8080/metrics&method=HEAD&codes=20a,403&interval=1m`,
			wantErr:  assert.Error,
		},
		{
			name:     "invalid interval",
			rawQuery: `target=http://localhost:8080/metrics&method=HEAD&codes=200&interval=zero`,
			wantErr:  assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := http.Request{
				URL: &url.URL{
					RawQuery: tt.rawQuery,
				},
			}

			req, err := ParseRequest(&r)
			tt.wantErr(t, err)
			assert.Equal(t, tt.wantReq, req)
		})
	}
}

func TestRequest_Encode(t *testing.T) {
	type fields struct {
		Target    string
		Method    string
		ValidCode []int
		Interval  time.Duration
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "full",
			fields: fields{
				Target:    "localhost:8080",
				Method:    http.MethodGet,
				ValidCode: []int{http.StatusOK, http.StatusForbidden},
				Interval:  time.Minute,
			},
			want: `codes=200%2C403&interval=1m0s&method=GET&target=localhost%3A8080`,
		},
		{
			name: "target only",
			fields: fields{
				Target: "localhost:8080",
			},
			want: `target=localhost%3A8080`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Request{
				Target:     tt.fields.Target,
				Method:     tt.fields.Method,
				ValidCodes: set.New(tt.fields.ValidCode...),
				Interval:   tt.fields.Interval,
			}
			assert.Equal(t, tt.want, r.Encode())
		})
	}
}

func TestRequest_LogValue(t *testing.T) {
	var output bytes.Buffer
	l := logtester.New(&output, slog.LevelInfo)

	req := Request{
		Target:     "http://localhost",
		Method:     http.MethodHead,
		ValidCodes: set.New(http.StatusOK),
		Interval:   time.Minute,
	}
	l.Info("request", "req", req)

	assert.Equal(t, `level=INFO msg=request req.target=http://localhost req.method=HEAD req.codes=[200] req.interval=1m0s
`, output.String())
}

func TestRequest_Equals(t *testing.T) {
	tests := []struct {
		name   string
		left   Request
		right  Request
		wantOK assert.BoolAssertionFunc
	}{
		{
			name:   "equal",
			left:   Request{Target: "http://localhost:8080", Method: http.MethodGet, ValidCodes: set.New(http.StatusOK), Interval: time.Hour},
			right:  Request{Target: "http://localhost:8080", Method: http.MethodGet, ValidCodes: set.New(http.StatusOK), Interval: time.Hour},
			wantOK: assert.True,
		},
		{
			name:   "different target",
			left:   Request{Target: "http://localhost:8080", Method: http.MethodGet, ValidCodes: set.New(http.StatusOK), Interval: time.Hour},
			right:  Request{Target: "http://localhost", Method: http.MethodGet, ValidCodes: set.New(http.StatusOK), Interval: time.Hour},
			wantOK: assert.False,
		},
		{
			name:   "different method",
			left:   Request{Target: "http://localhost:8080", Method: http.MethodGet, ValidCodes: set.New(http.StatusOK), Interval: time.Hour},
			right:  Request{Target: "http://localhost:8080", Method: http.MethodHead, ValidCodes: set.New(http.StatusOK), Interval: time.Hour},
			wantOK: assert.False,
		},
		{
			name:   "different statusCode",
			left:   Request{Target: "http://localhost:8080", Method: http.MethodGet, ValidCodes: set.New(http.StatusOK), Interval: time.Hour},
			right:  Request{Target: "http://localhost:8080", Method: http.MethodGet, ValidCodes: set.New(http.StatusTemporaryRedirect), Interval: time.Hour},
			wantOK: assert.False,
		},
		{
			name:   "different interval",
			left:   Request{Target: "http://localhost:8080", Method: http.MethodGet, ValidCodes: set.New(http.StatusOK), Interval: time.Hour},
			right:  Request{Target: "http://localhost:8080", Method: http.MethodGet, ValidCodes: set.New(http.StatusOK), Interval: time.Minute},
			wantOK: assert.False,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.wantOK(t, tt.left.Equals(tt.right))
		})
	}
}

package agent

import (
	"bytes"
	"flag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

var update = flag.Bool("update", false, "update golden images")

func TestConfiguration(t *testing.T) {
	tests := []struct {
		name  string
		input Configuration
	}{
		{
			name: "simple",
			input: Configuration{
				Monitor: "http://localhost:8080",
				Token:   "1234",
			},
		},
		{
			name: "global",
			input: Configuration{
				Monitor: "http://localhost:8080",
				Token:   "1234",
				Global: EndpointConfiguration{
					Interval:         time.Minute,
					Method:           http.MethodHead,
					ValidStatusCodes: []int{http.StatusOK, http.StatusUnauthorized},
				},
			},
		},
		{
			name: "hosts",
			input: Configuration{
				Monitor: "http://localhost:8080",
				Token:   "1234",
				Hosts: map[string]EndpointConfiguration{
					"http://localhost:8080": {
						Interval:         5 * time.Minute,
						Method:           http.MethodGet,
						ValidStatusCodes: []int{http.StatusOK},
					},
					"http://localhost:9090": {
						Skip: true,
					},
				},
			},
		},
		{
			name: "full",
			input: Configuration{
				Monitor: "http://localhost:8080",
				Token:   "1234",
				Global: EndpointConfiguration{
					Interval:         time.Minute,
					Method:           http.MethodHead,
					ValidStatusCodes: []int{http.StatusOK, http.StatusUnauthorized},
				},
				Hosts: map[string]EndpointConfiguration{
					"http://localhost:8080": {
						Interval:         5 * time.Minute,
						Method:           http.MethodGet,
						ValidStatusCodes: []int{http.StatusOK},
					},
					"http://localhost:9090": {
						Skip: true,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			output, err := yaml.Marshal(tt.input)
			fp := filepath.Join("testdata", t.Name()+".yaml")
			if *update {
				require.NoError(t, os.WriteFile(fp, output, 0644))
			}
			golden, err := os.ReadFile(fp)
			require.NoError(t, err)
			assert.Equal(t, string(golden), string(output))

			var read Configuration
			err = yaml.NewDecoder(bytes.NewReader(output)).Decode(&read)
			assert.NoError(t, err)
			assert.Equal(t, tt.input, read)
		})
	}
}

func TestConfiguration_Defaults(t *testing.T) {
	input := bytes.NewBufferString(`monitor: http://localhost:8080
token: "1234"
`)
	read, err := Load(input)
	require.NoError(t, err)

	want := Configuration{
		Monitor: "http://localhost:8080",
		Token:   "1234",
		Global:  DefaultGlobalConfiguration,
	}
	assert.Equal(t, want, read)
}

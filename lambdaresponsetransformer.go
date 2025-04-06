package traefik_lambdaresponsetransformer

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
)

type Config struct{}

func CreateConfig() *Config {
	return &Config{}
}

type LambdaResponseTransformer struct {
	next http.Handler
}

func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	return &LambdaResponseTransformer{next: next}, nil
}

type lambdaResponse struct {
	StatusCode      int               `json:"statusCode"`
	Body            string            `json:"body"`
	IsBase64Encoded bool              `json:"isBase64Encoded"`
	Headers         map[string]string `json:"headers"`
}

func (lrt *LambdaResponseTransformer) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	recorder := newResponseRecorder()
	lrt.next.ServeHTTP(recorder, req)

	// Parse the Lambda-style response body
	var lambdaResp lambdaResponse
	if err := json.Unmarshal(recorder.body.Bytes(), &lambdaResp); err != nil {
		http.Error(rw, "Invalid Lambda response format", http.StatusInternalServerError)
		return
	}

	// Set response headers from Lambda response
	for key, value := range lambdaResp.Headers {
		rw.Header().Set(key, value)
	}

	// Set CORS headers from incoming request headers if present
	if origin := req.Header.Get("Origin"); origin != "" {
		rw.Header().Set("Access-Control-Allow-Origin", origin)
	}
	if headers := req.Header.Get("Access-Control-Request-Headers"); headers != "" {
		rw.Header().Set("Access-Control-Allow-Headers", headers)
	}
	if method := req.Header.Get("Access-Control-Request-Method"); method != "" {
		rw.Header().Set("Access-Control-Allow-Methods", method)
	}

	// Write status code
	rw.WriteHeader(lambdaResp.StatusCode)

	// Decode body if needed
	var output []byte
	if lambdaResp.IsBase64Encoded {
		decoded, err := base64.StdEncoding.DecodeString(lambdaResp.Body)
		if err != nil {
			http.Error(rw, "Failed to decode base64 body", http.StatusInternalServerError)
			return
		}
		output = decoded
	} else {
		output = []byte(lambdaResp.Body)
	}

	rw.Write(output)
}

// responseRecorder captures the downstream response

type responseRecorder struct {
	body *bytes.Buffer
}

func newResponseRecorder() *responseRecorder {
	return &responseRecorder{
		body: &bytes.Buffer{},
	}
}

func (r *responseRecorder) Header() http.Header {
	return http.Header{}
}

func (r *responseRecorder) WriteHeader(statusCode int) {}

func (r *responseRecorder) Write(b []byte) (int, error) {
	return r.body.Write(b)
}

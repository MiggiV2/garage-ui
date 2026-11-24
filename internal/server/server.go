package server

import (
	"crypto/subtle"
	"encoding/json"
	"io/fs"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/MiggiV2/garage-ui/internal/config"
)

var allowedOperations = map[string]string{
	"GetClusterHealth": http.MethodGet,
	"GetClusterStatus": http.MethodGet,
	"ListBuckets":      http.MethodGet,
	"GetBucketInfo":    http.MethodGet,
	"CreateBucket":     http.MethodPost,
	"UpdateBucket":     http.MethodPost,
	"DeleteBucket":     http.MethodPost,
	"ListKeys":         http.MethodGet,
	"GetKeyInfo":       http.MethodGet,
	"CreateKey":        http.MethodPost,
	"UpdateKey":        http.MethodPost,
	"DeleteKey":        http.MethodPost,
	"AllowBucketKey":   http.MethodPost,
	"DenyBucketKey":    http.MethodPost,
}

func New(cfg config.Config, assets fs.FS) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/api/v2/", adminProxy(cfg))
	mux.Handle("GET /api/config", uiConfig(cfg))
	mux.Handle("/api/", http.NotFoundHandler())
	mux.Handle("/", spa(assets))

	root := http.NewServeMux()
	root.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	root.Handle("/", withBasicAuth(cfg, mux))
	return root
}

func adminProxy(cfg config.Config) http.Handler {
	target, err := url.Parse(cfg.AdminURL)
	if err != nil {
		panic("invalid GARAGE_ADMIN_URL: " + err.Error())
	}

	proxy := &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			r.SetURL(target)
			r.Out.URL.Path = "/v2/" + strings.TrimPrefix(r.In.URL.Path, "/api/v2/")
			r.Out.URL.RawPath = ""
			r.Out.Header.Del("Cookie")
			r.Out.Header.Set("Authorization", "Bearer "+cfg.AdminToken)
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			slog.Error("garage admin api unreachable", "operation", r.URL.Path, "error", err)
			http.Error(w, `{"message":"Garage admin API unreachable"}`, http.StatusBadGateway)
		},
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		operation := strings.TrimPrefix(r.URL.Path, "/api/v2/")
		method, ok := allowedOperations[operation]
		if !ok {
			http.NotFound(w, r)
			return
		}
		if r.Method != method {
			w.Header().Set("Allow", method)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		proxy.ServeHTTP(w, r)
	})
}

func uiConfig(cfg config.Config) http.Handler {
	body := struct {
		S3Endpoint string `json:"s3Endpoint"`
		S3Region   string `json:"s3Region"`
	}{cfg.S3Endpoint, cfg.S3Region}
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(body)
	})
}

func spa(assets fs.FS) http.Handler {
	files := http.FileServerFS(assets)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/")
		if name != "" {
			if info, err := fs.Stat(assets, name); err == nil && !info.IsDir() {
				if strings.HasPrefix(name, "assets/") {
					w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
				}
				files.ServeHTTP(w, r)
				return
			}
		}
		w.Header().Set("Cache-Control", "no-cache")
		http.ServeFileFS(w, r, assets, "index.html")
	})
}

func withBasicAuth(cfg config.Config, next http.Handler) http.Handler {
	if !cfg.BasicAuthEnabled() {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		userOK := subtle.ConstantTimeCompare([]byte(user), []byte(cfg.Username)) == 1
		passOK := subtle.ConstantTimeCompare([]byte(pass), []byte(cfg.Password)) == 1
		if !ok || !userOK || !passOK {
			w.Header().Set("WWW-Authenticate", `Basic realm="Garage UI", charset="UTF-8"`)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

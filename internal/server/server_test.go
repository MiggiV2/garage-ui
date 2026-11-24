package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/MiggiV2/garage-ui/internal/config"
)

type recordedRequest struct {
	method, path, query, auth, body string
}

func fakeGarage(t *testing.T) (*httptest.Server, *[]recordedRequest) {
	t.Helper()
	var requests []recordedRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		requests = append(requests, recordedRequest{r.Method, r.URL.Path, r.URL.RawQuery, r.Header.Get("Authorization"), string(body)})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTeapot)
		_, _ = w.Write([]byte(`{"from":"garage"}`))
	}))
	t.Cleanup(srv.Close)
	return srv, &requests
}

var assets = fstest.MapFS{
	"index.html":    {Data: []byte("<html>spa</html>")},
	"assets/app.js": {Data: []byte("console.log(1)")},
}

func newHandler(t *testing.T, cfg config.Config) http.Handler {
	t.Helper()
	return New(cfg, assets)
}

func baseConfig(adminURL string) config.Config {
	return config.Config{AdminURL: adminURL, AdminToken: "secret-token"}
}

func TestProxyForwardsAllowedOperationWithAdminToken(t *testing.T) {
	garage, requests := fakeGarage(t)
	h := newHandler(t, baseConfig(garage.URL))

	req := httptest.NewRequest(http.MethodPost, "/api/v2/UpdateBucket?id=abc", strings.NewReader(`{"quotas":null}`))
	req.Header.Set("Authorization", "Bearer client-supplied")
	req.Header.Set("X-Requested-With", "garage-ui")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusTeapot {
		t.Fatalf("expected upstream status to be passed through, got %d", rec.Code)
	}
	if rec.Body.String() != `{"from":"garage"}` {
		t.Fatalf("unexpected body %q", rec.Body.String())
	}
	if len(*requests) != 1 {
		t.Fatalf("expected 1 upstream request, got %d", len(*requests))
	}
	got := (*requests)[0]
	want := recordedRequest{http.MethodPost, "/v2/UpdateBucket", "id=abc", "Bearer secret-token", `{"quotas":null}`}
	if got != want {
		t.Fatalf("upstream request mismatch\n got: %+v\nwant: %+v", got, want)
	}
}

func TestProxyRejectsMutationsWithoutCSRFHeader(t *testing.T) {
	garage, requests := fakeGarage(t)
	h := newHandler(t, baseConfig(garage.URL))

	req := httptest.NewRequest(http.MethodPost, "/api/v2/DeleteBucket?id=abc", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "text/plain")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
	if len(*requests) != 0 {
		t.Fatal("cross-site style request must not reach garage")
	}
}

func TestProxyAllowsReadsWithoutCSRFHeader(t *testing.T) {
	garage, requests := fakeGarage(t)
	h := newHandler(t, baseConfig(garage.URL))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v2/ListBuckets", nil))
	if len(*requests) != 1 {
		t.Fatalf("expected GET to be forwarded, got %d upstream requests", len(*requests))
	}
}

func TestProxyRejectsOperationsOutsideAllowlist(t *testing.T) {
	garage, requests := fakeGarage(t)
	h := newHandler(t, baseConfig(garage.URL))

	for _, op := range []string{"UpdateClusterLayout", "ApplyClusterLayout", "CreateAdminToken"} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/v2/"+op, nil))
		if rec.Code != http.StatusNotFound {
			t.Errorf("%s: expected 404, got %d", op, rec.Code)
		}
	}
	if len(*requests) != 0 {
		t.Fatalf("expected no upstream requests, got %d", len(*requests))
	}
}

func TestProxyDoesNotForwardPathTraversal(t *testing.T) {
	garage, requests := fakeGarage(t)
	h := newHandler(t, baseConfig(garage.URL))

	for _, path := range []string{"/api/v2/../metrics", "/api/v2/ListBuckets/../../v1/metrics", "/api/v2/%2e%2e/metrics"} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code == http.StatusTeapot {
			t.Errorf("%s: request reached garage", path)
		}
	}
	if len(*requests) != 0 {
		t.Fatalf("expected no upstream requests, got %d: %+v", len(*requests), *requests)
	}
}

func TestProxyRejectsWrongMethod(t *testing.T) {
	garage, requests := fakeGarage(t)
	h := newHandler(t, baseConfig(garage.URL))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/v2/ListBuckets", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
	if len(*requests) != 0 {
		t.Fatal("wrong-method request must not reach garage")
	}
}

func TestBadGatewayWhenGarageUnreachable(t *testing.T) {
	h := newHandler(t, baseConfig("http://127.0.0.1:1"))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v2/GetClusterHealth", nil))
	if rec.Code != http.StatusBadGateway {
		t.Fatalf("expected 502, got %d", rec.Code)
	}
}

func TestServesStaticAssetsAndFallsBackToIndex(t *testing.T) {
	h := newHandler(t, baseConfig("http://unused"))

	cases := map[string]string{
		"/":              "<html>spa</html>",
		"/buckets/abc":   "<html>spa</html>",
		"/assets/app.js": "console.log(1)",
	}
	for path, want := range cases {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != http.StatusOK || rec.Body.String() != want {
			t.Errorf("%s: got %d %q, want 200 %q", path, rec.Code, rec.Body.String(), want)
		}
	}
}

func TestUIConfigExposesS3ConnectionDetailsButNoToken(t *testing.T) {
	cfg := baseConfig("http://unused")
	cfg.S3Endpoint, cfg.S3Region = "https://s3.example.test", "garage"
	h := newHandler(t, cfg)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/config", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	want := `{"s3Endpoint":"https://s3.example.test","s3Region":"garage"}`
	if strings.TrimSpace(rec.Body.String()) != want {
		t.Fatalf("got %s, want %s", rec.Body.String(), want)
	}
	if strings.Contains(rec.Body.String(), "secret-token") {
		t.Fatal("admin token leaked")
	}
}

func TestUnknownAPIPathsAreNotServedAsSPA(t *testing.T) {
	h := newHandler(t, baseConfig("http://unused"))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/nope", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestHealthzIsServedWithoutAuth(t *testing.T) {
	cfg := baseConfig("http://unused")
	cfg.Username, cfg.Password = "admin", "pw"
	h := newHandler(t, cfg)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestBasicAuthProtectsUIAndAPI(t *testing.T) {
	garage, _ := fakeGarage(t)
	cfg := baseConfig(garage.URL)
	cfg.Username, cfg.Password = "admin", "pw"
	h := newHandler(t, cfg)

	for _, path := range []string{"/", "/api/v2/ListBuckets"} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("%s without credentials: expected 401, got %d", path, rec.Code)
		}

		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.SetBasicAuth("admin", "wrong")
		rec = httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("%s with wrong password: expected 401, got %d", path, rec.Code)
		}

		req = httptest.NewRequest(http.MethodGet, path, nil)
		req.SetBasicAuth("admin", "pw")
		rec = httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code == http.StatusUnauthorized {
			t.Errorf("%s with valid credentials: got 401", path)
		}
	}
}

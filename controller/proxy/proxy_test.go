package proxy

import (
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestExtractAppName(t *testing.T) {
    tests := []struct {
        host    string
        want    string
        wantErr bool
    }{
        {"blog.1.2.3.4.nip.io", "blog", false},
        {"myapp.1.2.3.4.nip.io:80", "myapp", false},
        {"todo.192.168.1.100.nip.io", "todo", false},
        {"localhost", "localhost", false},
        {"", "", true},
    }
    for _, tt := range tests {
        got, err := extractAppName(tt.host)
        if (err != nil) != tt.wantErr {
            t.Errorf("extractAppName(%q) error = %v, wantErr %v", tt.host, err, tt.wantErr)
        }
        if got != tt.want {
            t.Errorf("extractAppName(%q) = %q, want %q", tt.host, got, tt.want)
        }
    }
}

func TestRouteTableThreadSafety(t *testing.T) {
    rt := NewRouteTable()
    done := make(chan struct{})
    go func() {
        for i := 0; i < 1000; i++ { rt.Register("app", "http://x") }
        close(done)
    }()
    for i := 0; i < 1000; i++ {
        rt.Lookup("app")
    }
    <-done
}

func TestProxyNotFound(t *testing.T) {
    rt := NewRouteTable()
    p := NewProxy(rt)
    req := httptest.NewRequest("GET", "/", nil)
    req.Host = "unknown.1.2.3.4.nip.io"
    rr := httptest.NewRecorder()
    p.ServeHTTP(rr, req)
    if rr.Code != http.StatusNotFound {
        t.Errorf("expected 404, got %d", rr.Code)
    }
}

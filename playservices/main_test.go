package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/AmitSuresh/playground/playservices/handlers"
	"go.uber.org/zap"
)

/*
test
go test -v
go test -coverprofile c.out
go tool cover -html c.out
go tool cover -html c.out -o coverage.html
go tool cover -func c.out
*/

func testHandler(t *testing.T, method, path string, body string, expectedBody string, handler http.HandlerFunc) {
	req, err := http.NewRequest(method, path, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	if !strings.Contains(rr.Body.String(), expectedBody) {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expectedBody)
	}
}

func TestWelcomeHandler(t *testing.T) {
	l, err := zap.NewProduction()
	if err != nil {
		t.Fatal(err)
	}
	testHandler(t, "GET", "/welcome", "", "Hello World!", handlers.NewWelcomeHandler(l).ServeHTTP)
}

func TestReadHandler(t *testing.T) {
	l, err := zap.NewProduction()
	if err != nil {
		t.Fatal(err)
	}
	testHandler(t, "POST", "/read", "test data", "reading!", handlers.NewReadHandler(l).ServeHTTP)
}

func TestMainFunction(t *testing.T) {
	go func() {
		main()
	}()
	resp, err := http.Get("http://localhost:9090/welcome")
	if err != nil {
		t.Fatalf("could not send GET request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status OK; got %v", resp.Status)
	}

	// Perform a POST request to the /read endpoint
	resp, err = http.Post("http://localhost:9090/read", "text/plain", strings.NewReader("test data"))
	if err != nil {
		t.Fatalf("could not send POST request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status OK; got %v", resp.Status)
	}
}

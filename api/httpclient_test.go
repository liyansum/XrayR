package api

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
)

func TestNewPanelHTTPClient(t *testing.T) {
	client := NewPanelHTTPClient(&Config{})
	if client.RetryCount != 3 {
		t.Fatalf("unexpected retry count: %d", client.RetryCount)
	}
	if client.GetClient().Timeout != 5*time.Second {
		t.Fatalf("unexpected default timeout: %s", client.GetClient().Timeout)
	}

	client = NewPanelHTTPClient(&Config{Timeout: 12})
	if client.GetClient().Timeout != 12*time.Second {
		t.Fatalf("unexpected configured timeout: %s", client.GetClient().Timeout)
	}
}

func TestRedactURL(t *testing.T) {
	redacted := RedactURL("https://panel.example/api?node_id=1&token=top-secret&muKey=second-secret")
	if strings.Contains(redacted, "top-secret") || strings.Contains(redacted, "second-secret") {
		t.Fatalf("credentials were not redacted: %s", redacted)
	}
	if !strings.Contains(redacted, "node_id=1") {
		t.Fatalf("non-sensitive query parameter was changed: %s", redacted)
	}
}

func TestRedactText(t *testing.T) {
	redacted := RedactText(`{"token":"top-secret","password":"hidden value","uuid":"user-credential","private_key":"private-value"} Authorization: Bearer abc.def`)
	for _, secret := range []string{"top-secret", "hidden value", "user-credential", "private-value", "abc.def"} {
		if strings.Contains(redacted, secret) {
			t.Fatalf("secret %q was not redacted: %s", secret, redacted)
		}
	}
}

func TestCheckResponseRedactsCredentials(t *testing.T) {
	response := (&resty.Response{RawResponse: &http.Response{StatusCode: http.StatusUnauthorized}}).
		SetBody([]byte(`{"token":"response-secret","message":"denied"}`))
	err := CheckResponse(response, "https://panel.example/api?token=query-secret", nil, 399)
	if err == nil {
		t.Fatal("expected an HTTP status error")
	}
	for _, secret := range []string{"response-secret", "query-secret"} {
		if strings.Contains(err.Error(), secret) {
			t.Fatalf("secret %q was not redacted: %s", secret, err)
		}
	}

	err = CheckResponse(nil, "https://panel.example/api", errors.New("token=transport-secret"), 399)
	if err == nil || strings.Contains(err.Error(), "transport-secret") {
		t.Fatalf("transport error was not safely redacted: %v", err)
	}
}

func TestReadLocalRuleList(t *testing.T) {
	path := filepath.Join(t.TempDir(), "rules.txt")
	if err := os.WriteFile(path, []byte("example\\.com\nblocked\\.test\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	rules := ReadLocalRuleList(path)
	if len(rules) != 2 || rules[0].ID != -1 || !rules[0].Pattern.MatchString("example.com") {
		t.Fatalf("unexpected rules: %#v", rules)
	}
}

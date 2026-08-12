package api

import (
	"bufio"
	"errors"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/bitly/go-simplejson"
	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
)

const (
	defaultPanelTimeout = 5 * time.Second
	maxLoggedBodySize   = 4096
)

var (
	sensitiveKeyPattern = regexp.MustCompile(`(?i)(["']?(?:access[_-]?token|refresh[_-]?token|token|api[_-]?key|private[_-]?key|server[_-]?key|key|muKey|authorization|password|passwd|secret|client[_-]?secret|uuid|psk)["']?\s*[:=]\s*)(?:"(?:\\.|[^"])*"|'(?:\\.|[^'])*'|[^&\s,}]+)`)
	bearerPattern       = regexp.MustCompile(`(?i)(bearer\s+)[A-Za-z0-9._~+/=-]+`)
)

// NewPanelHTTPClient creates the common HTTP client used by panel adapters.
func NewPanelHTTPClient(config *Config) *resty.Client {
	client := resty.New().SetRetryCount(3)
	if config.Timeout > 0 {
		client.SetTimeout(time.Duration(config.Timeout) * time.Second)
	} else {
		client.SetTimeout(defaultPanelTimeout)
	}
	client.OnError(func(_ *resty.Request, err error) {
		var responseErr *resty.ResponseError
		if errors.As(err, &responseErr) && responseErr.Err != nil {
			log.Print(RedactText(responseErr.Err.Error()))
			return
		}
		log.Print(RedactText(err.Error()))
	})
	return client
}

// EnableSafeDebug logs request metadata and a bounded response body without credentials.
func EnableSafeDebug(client *resty.Client) {
	client.SetDebug(false)
	client.OnBeforeRequest(func(_ *resty.Client, request *resty.Request) error {
		log.Printf("panel request: %s %s", request.Method, RedactURL(request.URL))
		return nil
	})
	client.OnAfterResponse(func(_ *resty.Client, response *resty.Response) error {
		log.Printf("panel response: %s %s status=%d body=%s",
			response.Request.Method,
			RedactURL(response.Request.URL),
			response.StatusCode(),
			SanitizeResponse(response),
		)
		return nil
	})
}

// ReadLocalRuleList reads regex rules shared by all panel adapters.
func ReadLocalRuleList(path string) []DetectRule {
	rules := make([]DetectRule, 0)
	if path == "" {
		return rules
	}

	file, err := os.Open(path)
	if err != nil {
		log.Printf("Error when opening rule file: %s", err)
		return rules
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("Error when closing rule file: %s", err)
		}
	}()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		rules = append(rules, DetectRule{ID: -1, Pattern: regexp.MustCompile(scanner.Text())})
	}
	if err := scanner.Err(); err != nil {
		log.Printf("Error while reading rule file: %s", err)
	}
	return rules
}

// CheckResponse validates transport and HTTP status while keeping logs credential-safe.
func CheckResponse(response *resty.Response, requestURL string, requestErr error, maxSuccessStatus int) error {
	requestURL = RedactURL(requestURL)
	if requestErr != nil {
		return fmt.Errorf("request %s failed: %s", requestURL, RedactText(requestErr.Error()))
	}
	if response == nil {
		return fmt.Errorf("request %s failed: empty response", requestURL)
	}
	if response.StatusCode() > maxSuccessStatus {
		return fmt.Errorf("request %s failed with status %d: %s", requestURL, response.StatusCode(), SanitizeResponse(response))
	}
	return nil
}

// ParseJSONResponse validates and parses responses used by JSON-based adapters.
func ParseJSONResponse(response *resty.Response, requestURL string, requestErr error, maxSuccessStatus int) (*simplejson.Json, error) {
	if err := CheckResponse(response, requestURL, requestErr, maxSuccessStatus); err != nil {
		return nil, err
	}
	result, err := simplejson.NewJson(response.Body())
	if err != nil {
		return nil, fmt.Errorf("response from %s is invalid JSON: %s", RedactURL(requestURL), SanitizeResponse(response))
	}
	return result, nil
}

// RedactURL masks credentials stored in URL query parameters.
func RedactURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return RedactText(rawURL)
	}
	query := parsed.Query()
	for key := range query {
		if isSensitiveKey(key) {
			query.Set(key, "***")
		}
	}
	parsed.RawQuery = query.Encode()
	if parsed.User != nil {
		if _, hasPassword := parsed.User.Password(); hasPassword {
			parsed.User = url.UserPassword(parsed.User.Username(), "***")
		}
	}
	return parsed.String()
}

// RedactText masks common credential fields in errors and response bodies.
func RedactText(text string) string {
	text = bearerPattern.ReplaceAllString(text, "${1}***")
	return sensitiveKeyPattern.ReplaceAllString(text, `${1}"***"`)
}

// SanitizeResponse returns a redacted, bounded response body for diagnostics.
func SanitizeResponse(response *resty.Response) string {
	if response == nil {
		return ""
	}
	body := RedactText(string(response.Body()))
	if len(body) > maxLoggedBodySize {
		body = body[:maxLoggedBodySize] + "...(truncated)"
	}
	return body
}

func isSensitiveKey(key string) bool {
	normalized := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(key, "_", ""), "-", ""))
	switch normalized {
	case "accesstoken", "refreshtoken", "token", "apikey", "privatekey", "serverkey", "key", "mukey", "authorization", "password", "passwd", "secret", "clientsecret", "uuid", "psk":
		return true
	default:
		return false
	}
}

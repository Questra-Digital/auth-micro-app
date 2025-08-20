package api

import (
	"api-gateway/config"
	"api-gateway/utils"
	"bytes"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// ResourceClient handles HTTP requests to Resource service
type ResourceClient struct {
	client *http.Client
}

// NewResourceClient creates a new Resource client with default timeout
func NewResourceClient() *ResourceClient {
	return &ResourceClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ForwardToResourceAPI forwards the request to the Resource API
func (rc *ResourceClient) ForwardToResourceAPI(c *gin.Context, claims *utils.CustomClaims) error {
	// Get request body
	var body []byte
	if c.Request.Body != nil {
		body, _ = io.ReadAll(c.Request.Body)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
	}

	// Create new request to Resource API
	resourceURL := config.AppConfig.ResourceServiceURL + c.Request.URL.Path
	if c.Request.URL.RawQuery != "" {
		resourceURL += "?" + c.Request.URL.RawQuery
	}

	req, err := http.NewRequest(c.Request.Method, resourceURL, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	// Copy headers
	for key, values := range c.Request.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// Add user context headers
	req.Header.Set("X-User-ID", claims.UserID)
	req.Header.Set("X-User-Email", claims.Email)
	req.Header.Set("X-User-Scopes", strings.Join(claims.Scopes, ","))

	// Make request to Resource API
	resp, err := rc.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}

	// Set response status and body
	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)

	return nil
}

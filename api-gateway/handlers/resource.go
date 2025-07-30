package handlers

import (
	"api-gateway/api"
	"api-gateway/models"
	"api-gateway/redis"
	"api-gateway/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Add the missing constant
const ActionResourceAccess models.EventAction = "RESOURCE_ACCESS"

// ResourceHandler handles all resource-related requests
func ResourceHandler(c *gin.Context) {
	log := utils.NewLogger()
	resourceClient := api.NewResourceClient()
	authClient := api.NewAuthClient()

	// Extract request context info
	reqCtx := models.RequestContext{
		IP:     c.ClientIP(),
		Method: c.Request.Method,
		Path:   c.FullPath(),
	}

	// 1. Extract sessionId from cookie
	sessionID, err := c.Cookie("sessionId")
	if err != nil {
		log.Warn("Missing sessionId cookie")

		msg := "Missing session ID"
		auditEntry := log.NewAuditEntry(
			models.EventGroupAuth,
			ActionResourceAccess,
			nil,
			nil,
			reqCtx,
			http.StatusUnauthorized,
			&msg,
		)
		log.LogAuditEntry(auditEntry)

		c.JSON(http.StatusUnauthorized, gin.H{"error": msg})
		return
	}

	// 2. Get access token from Redis using sessionId
	sessionData, err := redis.GetSessionData(sessionID)
	if err != nil {
		log.Warn("Failed to get session data: %v", err)

		msg := "Invalid session"
		auditEntry := log.NewAuditEntry(
			models.EventGroupAuth,
			ActionResourceAccess,
			nil,
			nil,
			reqCtx,
			http.StatusUnauthorized,
			&msg,
		)
		log.LogAuditEntry(auditEntry)

		c.JSON(http.StatusUnauthorized, gin.H{"error": msg})
		return
	}

	// Check if access token exists in session data
	accessToken, exists := sessionData["token"]
	if !exists || accessToken == "" {
		log.Warn("No access token found in session")

		msg := "No access token found"
		auditEntry := log.NewAuditEntry(
			models.EventGroupAuth,
			ActionResourceAccess,
			nil,
			nil,
			reqCtx,
			http.StatusUnauthorized,
			&msg,
		)
		log.LogAuditEntry(auditEntry)

		c.JSON(http.StatusUnauthorized, gin.H{"error": msg})
		return
	}

	// 3. Validate JWT token
	claims, err := utils.ValidateJWTToken(accessToken)
	if err != nil {
		// Check if token is expired by checking the error message
		if err.Error() == "token has expired" || err.Error() == "Token is expired" {
			log.Info("Access token expired, attempting to refresh")

			// Try to refresh the token
			newAccessToken, err := authClient.RefreshAccessToken(sessionData, log)
			if err != nil {
				log.Warn("Failed to refresh access token: %v", err)

				// Delete session data since refresh failed
				redis.DeleteSession(sessionID)

				msg := "Session expired, please login again"
				auditEntry := log.NewAuditEntry(
					models.EventGroupAuth,
					ActionResourceAccess,
					nil,
					nil,
					reqCtx,
					http.StatusUnauthorized,
					&msg,
				)
				log.LogAuditEntry(auditEntry)

				c.JSON(http.StatusUnauthorized, gin.H{"error": msg})
				return
			}

			// Update session data with new access token
			if err := redis.UpdateSessionField(sessionID, "token", newAccessToken); err != nil {
				log.Error("Failed to update session with new access token: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update session"})
				return
			}

			// Validate the new token
			claims, err = utils.ValidateJWTToken(newAccessToken)
			if err != nil {
				log.Warn("New access token validation failed: %v", err)
				msg := "Invalid access token"
				auditEntry := log.NewAuditEntry(
					models.EventGroupAuth,
					ActionResourceAccess,
					nil,
					nil,
					reqCtx,
					http.StatusUnauthorized,
					&msg,
				)
				log.LogAuditEntry(auditEntry)

				c.JSON(http.StatusUnauthorized, gin.H{"error": msg})
				return
			}
		} else {
			log.Warn("JWT token validation failed: %v", err)

			msg := "Invalid access token"
			auditEntry := log.NewAuditEntry(
				models.EventGroupAuth,
				ActionResourceAccess,
				nil,
				nil,
				reqCtx,
				http.StatusUnauthorized,
				&msg,
			)
			log.LogAuditEntry(auditEntry)

			c.JSON(http.StatusUnauthorized, gin.H{"error": msg})
			return
		}
	}

	// 4. Check required scopes based on HTTP method
	requiredScope := getRequiredScope(c.Request.Method)
	if !hasRequiredScope(claims.Scopes, requiredScope) {
		log.Warn("Insufficient scopes. Required: %s, User scopes: %v", requiredScope, claims.Scopes)

		msg := "Insufficient permissions"
		auditEntry := log.NewAuditEntry(
			models.EventGroupAuth,
			ActionResourceAccess,
			&claims.UserID,
			&claims.Email,
			reqCtx,
			http.StatusForbidden,
			&msg,
		)
		log.LogAuditEntry(auditEntry)

		c.JSON(http.StatusForbidden, gin.H{"error": msg})
		return
	}

	// 5. Forward request to Resource API
	err = resourceClient.ForwardToResourceAPI(c, claims)
	if err != nil {
		log.Error("Failed to forward request to Resource API: %v", err)

		msg := "Resource service unavailable"
		auditEntry := log.NewAuditEntry(
			models.EventGroupAuth,
			ActionResourceAccess,
			&claims.UserID,
			&claims.Email,
			reqCtx,
			http.StatusInternalServerError,
			&msg,
		)
		log.LogAuditEntry(auditEntry)

		c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}

	// Log successful access
	auditEntry := log.NewAuditEntry(
		models.EventGroupAuth,
		ActionResourceAccess,
		&claims.UserID,
		&claims.Email,
		reqCtx,
		http.StatusOK,
		nil,
	)
	log.LogAuditEntry(auditEntry)
}

// getRequiredScope returns the required scope based on HTTP method
func getRequiredScope(method string) string {
	switch method {
	case "GET":
		return "read"
	case "POST", "PUT", "DELETE":
		return "write"
	default:
		return "read"
	}
}

// hasRequiredScope checks if the user has the required scope
func hasRequiredScope(userScopes []string, requiredScope string) bool {
	for _, scope := range userScopes {
		if scope == requiredScope {
			return true
		}
	}
	return false
}

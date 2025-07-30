package handlers

import (
	"net/http"
	"resource-service/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// In-memory storage for resources
var resources = make(map[string]models.Resource)

// populateDummyData creates and adds dummy resources to the map
func populateDummyData() {
	// Create dummy resources
	dummyResources := []models.Resource{
		{
			ID:          "550e8400-e29b-41d4-a716-446655440001",
			Name:        "User Management System",
			Description: "A comprehensive system for managing user accounts, roles, and permissions",
			CreatedAt:   time.Now().Add(-24 * time.Hour),
			UpdatedAt:   time.Now().Add(-2 * time.Hour),
		},
		{
			ID:          "550e8400-e29b-41d4-a716-446655440002",
			Name:        "Payment Gateway",
			Description: "Secure payment processing system with multiple payment method support",
			CreatedAt:   time.Now().Add(-48 * time.Hour),
			UpdatedAt:   time.Now().Add(-6 * time.Hour),
		},
		{
			ID:          "550e8400-e29b-41d4-a716-446655440003",
			Name:        "Analytics Dashboard",
			Description: "Real-time analytics and reporting dashboard with interactive charts",
			CreatedAt:   time.Now().Add(-72 * time.Hour),
			UpdatedAt:   time.Now().Add(-12 * time.Hour),
		},
		{
			ID:          "550e8400-e29b-41d4-a716-446655440004",
			Name:        "Email Service",
			Description: "Reliable email delivery service with template management and tracking",
			CreatedAt:   time.Now().Add(-96 * time.Hour),
			UpdatedAt:   time.Now().Add(-1 * time.Hour),
		},
		{
			ID:          "550e8400-e29b-41d4-a716-446655440005",
			Name:        "File Storage System",
			Description: "Cloud-based file storage with version control and sharing capabilities",
			CreatedAt:   time.Now().Add(-120 * time.Hour),
			UpdatedAt:   time.Now().Add(-4 * time.Hour),
		},
		{
			ID:          "550e8400-e29b-41d4-a716-446655440006",
			Name:        "Notification Service",
			Description: "Multi-channel notification system supporting push, SMS, and email",
			CreatedAt:   time.Now().Add(-144 * time.Hour),
			UpdatedAt:   time.Now().Add(-8 * time.Hour),
		},
		{
			ID:          "550e8400-e29b-41d4-a716-446655440007",
			Name:        "API Gateway",
			Description: "Centralized API management with rate limiting and authentication",
			CreatedAt:   time.Now().Add(-168 * time.Hour),
			UpdatedAt:   time.Now().Add(-3 * time.Hour),
		},
		{
			ID:          "550e8400-e29b-41d4-a716-446655440008",
			Name:        "Database Management",
			Description: "Database administration tools with backup and recovery features",
			CreatedAt:   time.Now().Add(-192 * time.Hour),
			UpdatedAt:   time.Now().Add(-5 * time.Hour),
		},
	}

	// Add dummy resources to the map
	for _, resource := range dummyResources {
		resources[resource.ID] = resource
	}
}

// init function to populate dummy data when the package is imported
func init() {
	populateDummyData()
}

// GetAllResources returns all resources
func GetAllResources(c *gin.Context) {
	// Extract user context from headers (set by API Gateway)
	userID := c.GetHeader("X-User-ID")
	userEmail := c.GetHeader("X-User-Email")
	userScopes := c.GetHeader("X-User-Scopes")

	// Convert map to slice for response
	var resourceList []models.Resource
	for _, resource := range resources {
		resourceList = append(resourceList, resource)
	}

	c.JSON(http.StatusOK, gin.H{
		"resources":   resourceList,
		"user_id":     userID,
		"user_email":  userEmail,
		"user_scopes": userScopes,
	})
}

// GetResource returns a specific resource by ID
func GetResource(c *gin.Context) {
	id := c.Param("id")

	resource, exists := resources[id]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Resource not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"resource":    resource,
		"user_id":     c.GetHeader("X-User-ID"),
		"user_email":  c.GetHeader("X-User-Email"),
		"user_scopes": c.GetHeader("X-User-Scopes"),
	})
}

// CreateResource creates a new resource
func CreateResource(c *gin.Context) {
	var req models.CreateResourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Generate unique ID
	id := uuid.New().String()
	now := time.Now()

	resource := models.Resource{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	resources[id] = resource

	c.JSON(http.StatusCreated, gin.H{
		"resource":    resource,
		"user_id":     c.GetHeader("X-User-ID"),
		"user_email":  c.GetHeader("X-User-Email"),
		"user_scopes": c.GetHeader("X-User-Scopes"),
	})
}

// UpdateResource updates an existing resource
func UpdateResource(c *gin.Context) {
	id := c.Param("id")

	// Check if resource exists
	existingResource, exists := resources[id]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Resource not found"})
		return
	}

	var req models.UpdateResourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Update fields if provided
	if req.Name != "" {
		existingResource.Name = req.Name
	}
	if req.Description != "" {
		existingResource.Description = req.Description
	}
	existingResource.UpdatedAt = time.Now()

	resources[id] = existingResource

	c.JSON(http.StatusOK, gin.H{
		"resource":    existingResource,
		"user_id":     c.GetHeader("X-User-ID"),
		"user_email":  c.GetHeader("X-User-Email"),
		"user_scopes": c.GetHeader("X-User-Scopes"),
	})
}

// DeleteResource deletes a resource
func DeleteResource(c *gin.Context) {
	id := c.Param("id")

	// Check if resource exists
	_, exists := resources[id]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Resource not found"})
		return
	}

	// Delete resource
	delete(resources, id)

	c.JSON(http.StatusOK, gin.H{
		"message":     "Resource deleted successfully",
		"user_id":     c.GetHeader("X-User-ID"),
		"user_email":  c.GetHeader("X-User-Email"),
		"user_scopes": c.GetHeader("X-User-Scopes"),
	})
}

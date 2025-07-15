package handlers

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"otp-service/config"
	"otp-service/models"
	"otp-service/redis"
	"otp-service/utils"
	"time"
)

func GenerateOTPHandler(c *gin.Context) {
	// STEP 1: Get or generate session ID
	sessionID, err := c.Cookie("session_id")
	var session *models.OTPSession

	if err != nil || sessionID == "" {
		// Generate new session ID if cookie doesn't exist or is empty
		sessionID = utils.GenerateSessionID()
		// Set session ID as a cookie
		c.SetCookie("session_id", sessionID, int(config.OTPTTL.Seconds()), "/", "", false, true)
		// Create new session
		session = &models.OTPSession{
			SessionID: sessionID,
			Resends:   0,
			Attempts:  0,
			CreatedAt: time.Now(),
		}
	} else {
		// Retrieve existing session (if any)
		sessionValue, err := redis.GetSession(sessionID)
		if err != nil {
			if err.Error() == "redis: nil" {
				// Session doesn't exist
				c.JSON(http.StatusBadRequest, gin.H{"error": "Session against this sessionId does not exist"})
				return
			} else {
				// Other Redis error
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Redis error while retrieving session"})
				return
			}
		}

		// Session exists, use it
		session = &sessionValue
		
		// Check if maximum resends exceeded
		if session.Resends >= config.MaxResends {
			redis.DeleteSession(sessionID)
			c.SetCookie("session_id", "", -1, "/", "", false, true) // Expire cookie
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Maximum OTP resends exceeded"})
			return
		}
		session.Resends++
	}

	// STEP 4: Generate and hash OTP
	otp := utils.GenerateSecureOTP(config.OTPLength)
	hashedOTP, err := utils.HashOTP(otp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash OTP"})
		return
	}

	// STEP 5: Update session data
	session.SessionID = sessionID
	session.OTPHash = hashedOTP
	session.CreatedAt = time.Now()

	// STEP 6: Store session in Redis (dereference pointer to pass value)
	if err := redis.StoreSession(*session, config.OTPTTL); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store session in Redis"})
		return
	}

	// STEP 7: Respond
	log.Printf("OTP generated for session: %s (OTP: %s)", sessionID, otp)
	c.JSON(http.StatusOK, gin.H{"success": "OTP generated successfully"})
}

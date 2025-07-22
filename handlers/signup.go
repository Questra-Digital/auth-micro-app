package handlers

import (
	"auth-microservice/utils"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"time"
	"fmt"
	"github.com/go-playground/validator/v10"
)

type SignupRequest struct {
	Email string `json:"email" validate:"required,email"`
}

var validate = validator.New()

func SignupHandler(w http.ResponseWriter, r *http.Request) {

	//A variable of the struct
	var req SignupRequest

	//Email field must be present
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Email == "" {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	//Validate through valdiator that it is a valid email
	if err := validate.Struct(req); err != nil {
		http.Error(w, "Validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	code := generateCode()
	//Save the code as a value to the key email to a dummy storage.
	utils.SaveVerificationCode(req.Email, code)
	log.Printf("Generated code %s for %s", code, req.Email)

	//Send verification code to the mentioned email
	if err := utils.SendVerificationEmail(req.Email, code); err != nil {
		log.Printf("Failed to send email: %v", err)
		http.Error(w, "Failed to send verification email", http.StatusInternalServerError)
		return
	}

	//Successful message returned
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := Response{
        Message: "Verification code sent",
        Success: true,
    }
    json.NewEncoder(w).Encode(response)
}

func generateCode() string {
    r := rand.New(rand.NewSource(time.Now().UnixNano()))
    return fmt.Sprintf("%06d", r.Intn(1000000))
}
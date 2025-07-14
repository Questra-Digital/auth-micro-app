package handlers

import (
	"auth-microservice/utils"
	"encoding/json"
	"net/http"
)

type VerifyRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

type Response struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

func VerifyHandler(w http.ResponseWriter, r *http.Request) {
	//Declare a variable of VerifyRequest struct
	var req VerifyRequest

	//Decode the incoming request to struct, email and code must not be empty
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Email == "" || req.Code == "" {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	//Verify that code against the given email exists.
	if utils.VerifyCode(req.Email, req.Code) {
		//Send success message
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := Response{
			Message: "Email verification successful",
			Success: true,
		}

		json.NewEncoder(w).Encode(response)

		//Perhaps send a signed token as a cookie after success.
		
	} else {
		http.Error(w, "Invalid or expired code", http.StatusUnauthorized)
	}
}

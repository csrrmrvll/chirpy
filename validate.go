package main

import (
	"encoding/json"
	"net/http"
)

func (cfg *apiConfig) handlerValidateChirp(w http.ResponseWriter, r *http.Request) {
	type RequestParameters struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	var req RequestParameters
	if err := decoder.Decode(&req); err != nil {
		responseWithError(w, "Something went wrong")
		return
	}

	if len(req.Body) > 140 {
		responseWithError(w, "Chirp is too long")
		return
	}
	responseWithJSON(w, true)
}

func responseWithJSON(w http.ResponseWriter, valid bool) {

	type JSONResponse struct {
		Valid bool `json:"valid"`
	}

	response := JSONResponse{
		Valid: valid,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(response)
}

func responseWithError(w http.ResponseWriter, message string) {
	type ErrorResponse struct {
		Error string `json:"error"`
	}

	response := ErrorResponse{
		Error: message,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(response)
}

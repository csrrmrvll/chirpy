package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

func handlerChirpsValidate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	const maxChirpLength = 140
	if len(params.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}

	responseBody := replaceBadWords(params.Body)
	type returnVals struct {
		Cleaned_body string `json:"cleaned_body"`
	}

	respondWithJSON(w, http.StatusOK, returnVals{
		Cleaned_body: responseBody,
	})
}

func replaceBadWords(input string) string {
	notAllowedSubstrings := []string{"kerfuffle", "sharbert", "fornax"}
	for _, bad := range notAllowedSubstrings {
		for _, word := range strings.Fields(input) {
			if strings.EqualFold(word, bad) {
				input = strings.ReplaceAll(input, word, "****")
			}
		}
	}
	return input
}

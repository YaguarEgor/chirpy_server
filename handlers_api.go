package main

import (
	"encoding/json"
	"net/http"
	"slices"
	"strings"
)

func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}


func handlerValidation(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	type parameters struct {
		Body string `json:"body"`
	}
	type returnVals struct {
		CleanedBody string `json:"cleaned_body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode request", err)
		return
	}
	if len(params.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return 
	}
	respondWithJSON(w, http.StatusOK, returnVals {
		CleanedBody: removeBadWords(params.Body),
	})
}



func removeBadWords(s string) string {
	bad_words := []string{"profane", "kerfuffle", "sharbert", "fornax"}
	words := strings.Split(s, " ")
	for ind, word := range words {
		if slices.Contains(bad_words, strings.ToLower(word)) {
			words[ind] = "****"
		}
	}
	return strings.Join(words, " ")
}
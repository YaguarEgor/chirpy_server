package main

import (
	"encoding/json"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/YaguarEgor/chirpy_server/internal/auth"
	"github.com/YaguarEgor/chirpy_server/internal/database"
	"github.com/google/uuid"
)


type User struct {
	ID        	 uuid.UUID 	`json:"id"`
	CreatedAt 	 time.Time 	`json:"created_at"`
	UpdatedAt 	 time.Time 	`json:"updated_at"`
	Email     	 string    	`json:"email"`
}

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	UserID    uuid.UUID `json:"user_id"`
	Body      string    `json:"body"`
}

func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
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

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	type parameters struct {
		Email 	 	string 	`json:"email"`
		Password 	string 	`json:"password"`
	}

	var params parameters

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode request", err)
		return
	}

	hashed_passwd, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't hash password", err)
		return
	}

	user, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
		Email: params.Email,
		HashedPassword: hashed_passwd,
	})

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create user", err)
		return
	}
	my_user := User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}
	respondWithJSON(w, 201, my_user)
}

func (cfg *apiConfig) handlerChirps(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	type parameters struct {
		Body string `json:"body"`
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get JWT token", err)
		return
	}
	id, err := auth.ValidateJWT(token, cfg.secret_key)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT token", err)
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode request", err)
		return
	}
	if len(params.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}

	chirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams {
		Body:  removeBadWords(params.Body),
		UserID: id,
	})

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't add chirp to database", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, Chirp {
		ID: chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body: chirp.Body,
		UserID: chirp.UserID,
	})
	
}

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	chirps, err := cfg.db.GetChirps(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get chirps", err)
		return
	}

	data := []Chirp {}
	for _, chirp := range chirps {
		data = append(data, Chirp {
			ID: chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body: chirp.Body,
			UserID: chirp.UserID,
		})
	}

	respondWithJSON(w, http.StatusOK, data)
}

func (cfg *apiConfig) handlerGetChirp(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	chirp_id := r.PathValue("chirpID")
	chirp, err := cfg.db.GetChirp(r.Context(), uuid.MustParse(chirp_id))
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Couldn't get chirp", err)
		return
	}

	respondWithJSON(w, http.StatusOK, Chirp {
		ID: chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body: chirp.Body,
		UserID: chirp.UserID,
	})
}

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	type parameters struct {
		Email 	 	string 	`json:"email"`
		Password 	string 	`json:"password"`
	}

	type response struct {
		User
		JWTToken  	 string 	`json:"token"`
		RefreshToken string 	`json:"refresh_token"`
	}

	var params parameters
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode request", err)
		return
	}

	user, err := cfg.db.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't find user with that email", err)
		return
	}

	err = auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}

	jwt_token, err := auth.MakeJWT(user.ID, cfg.secret_key, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error when creating JWT token", err)
		return
	}

	refresh_token, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error when creating Refresh token", err)
		return
	}

	_, err = cfg.db.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token: refresh_token,
		UserID: user.ID,
		ExpiresAt: time.Now().UTC().Add(time.Hour * 24 * 60),
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error when adding Refresh token to database", err)
		return
	}

	respondWithJSON(w, http.StatusOK, response {
		User: User {
			ID: user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email: user.Email,
		},
		JWTToken: jwt_token,
		RefreshToken: refresh_token,
	})

}

func (cfg *apiConfig) handlerRefreshToken(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Token string `json:"token"`
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get token", err)
		return
	} 

	user, err := cfg.db.GetUserFromRefreshToken(r.Context(), token)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't get refresh token", err)
		return
	} 

	access_token, err := auth.MakeJWT(user.ID, cfg.secret_key, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't create access token", err)
		return
	} 

	respondWithJSON(w, http.StatusOK, response {
		Token: access_token,
	})
}

func (cfg *apiConfig) handlerRevokeToken(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get token", err)
		return
	} 
	_, err = cfg.db.RevokeRefreshToken(r.Context(), token)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't revoke token", err)
		return
	} 
	respondWithJSON(w, http.StatusNoContent, nil)
}
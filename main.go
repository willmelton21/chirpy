package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/willmelton21/chirpy/internal/auth"
	"github.com/willmelton21/chirpy/internal/database"

	_ "github.com/lib/pq"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	Token     string    `json:"Token"`
	Is_Chirpy_Red bool  `json:"is_chirpy_red"`
}

type LoginRequest struct {
	Email            string `json:"email"`
	Password         string `json:"password"`
}

type apiConfig struct {
	fileserverHits atomic.Int32
	dbs            *database.Queries
	Platform       string
	Secret         string
}

type parameters struct {
	Body string `json:"body"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
type cleanedBody struct {
	Body string `json:"cleaned_body"`
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) UpgradeUser(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		User_id string `json:"user_id"`
	}

	type upgradeParams struct {
		Event string `json:"event"`
		Data  Data   `json:"data"`   
	}

	key, err := auth.GetAPIKey(r.Header)
	if err != nil {
		respondWithError(w,401,"Couldn't GetApi key from  header")
		return
	}

	if key != os.Getenv("POLKA_KEY") {
		respondWithError(w,401,"Key does not match .env key")
		return


	}
	
	decoder := json.NewDecoder(r.Body)
   var upgradeStruct upgradeParams
	err = decoder.Decode(&upgradeStruct)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	if upgradeStruct.Event != "user.upgraded" {
		respondWithJSON(w,204,"")
		return
	}
	parsedID, err := uuid.Parse(upgradeStruct.Data.User_id)
	err = cfg.dbs.Upgradeuser(r.Context(),uuid.UUID(parsedID))
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Couldn't Find Upgrade User")
		return
	}


	respondWithJSON(w,204,"")
	return
}

func (cfg *apiConfig) DeleteChirp(w http.ResponseWriter, r *http.Request) {
  	authHeader := r.Header.Get("Authorization")	
	if authHeader == ""{
		respondWithError(w, http.StatusUnauthorized, "Header auth token was empty")
		return
	}
	splitAuth := strings.Split(authHeader, " ")
	if len(splitAuth) < 2 || splitAuth[0] != "Bearer" {
		respondWithError(w, http.StatusBadRequest, "malformed authorization header")
		return
	}

	token := splitAuth[1]

   tokenSecret := os.Getenv("SECRET")

	authedUserID, err := auth.ValidateJWT(token,tokenSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't Get user from token")
		return
	}

	currID := r.PathValue("chirpID")
	chirpID, err := uuid.Parse(currID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

   chirp,err := cfg.dbs.GetChirpByID(r.Context(), chirpID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
   if chirp.UserID != authedUserID {
      respondWithError(w,403,"Chirp User ID did not match authorized user ID")
      return
   }

   err = cfg.dbs.DeleteChirpByID(r.Context(), chirpID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	respondWithJSON(w,204,"")
}

func (cfg *apiConfig) UpdateUserInfo(w http.ResponseWriter, r *http.Request) {

	type parameters struct {
		Password         string `json:"password"`
		Email            string `json:"email"`
	}

	authHeader := r.Header.Get("Authorization")	
	if authHeader == ""{
		respondWithError(w, http.StatusUnauthorized, "Header auth token was empty")
		return
	}
	splitAuth := strings.Split(authHeader, " ")
	if len(splitAuth) < 2 || splitAuth[0] != "Bearer" {
		respondWithError(w, http.StatusBadRequest, "malformed authorization header")
		return
	}

	token := splitAuth[1]

	userID, err := auth.GetUserIDFromToken(token)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't Get user from token")
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "couldn't hash password")
		return
	}

	updatedUser, err := cfg.dbs.UpdateEmailAndPass(r.Context(),database.UpdateEmailAndPassParams{Email: params.Email, HashedPassword: hashedPassword,ID: userID})
	if err != nil {
		respondWithError(w,http.StatusInternalServerError, "Couldn't update credentials")
	}

	userStruct := User{
			ID:        updatedUser.ID,
			CreatedAt: updatedUser.CreatedAt,
			UpdatedAt: updatedUser.UpdatedAt,
			Email:     updatedUser.Email,
		   Is_Chirpy_Red: updatedUser.IsChirpyRed.Bool,
		}

	respondWithJSON(w, http.StatusOK,userStruct)

}

func (cfg *apiConfig) Revoke(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")	
	if authHeader == ""{
		respondWithError(w, http.StatusBadRequest, "Header auth token was empty")
		return
	}
	splitAuth := strings.Split(authHeader, " ")
	if len(splitAuth) < 2 || splitAuth[0] != "Bearer" {
		respondWithError(w, http.StatusBadRequest, "malformed authorization header")
		return
	}

	token := splitAuth[1]

	 err := cfg.dbs.RevokeToken(r.Context(),token)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Token couldn't be revoked")
		return
	}	
	respondWithJSON(w,204,"")
	return
}

func (cfg *apiConfig) Refresh(w http.ResponseWriter, r *http.Request) {

	type tokenResponse struct {
		Token string `json:"token"`
	}

	authHeader := r.Header.Get("Authorization")	
	if authHeader == ""{
		respondWithError(w, http.StatusBadRequest, "Header auth token was empty")
		return
	}
	splitAuth := strings.Split(authHeader, " ")
	if len(splitAuth) < 2 || splitAuth[0] != "Bearer" {
		respondWithError(w, http.StatusBadRequest, "malformed authorization header")
		return
	}

	token := splitAuth[1]

	validUser, err := cfg.dbs.GetUserFromRefreshToken(r.Context(),token)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Token not in database or expired")
		return
	}

	accessToken, err := auth.MakeJWT(validUser.ID,cfg.Secret,time.Hour)



	tokenStruct := tokenResponse{Token: accessToken}

	respondWithJSON(w,http.StatusOK,tokenStruct)


}


func (cfg *apiConfig) Login(w http.ResponseWriter, r *http.Request) {
type parameters struct {
		Password         string `json:"password"`
		Email            string `json:"email"`
	}
	type response struct {
		User
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	user, err := cfg.dbs.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}

	err = auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password",)
		return
	}

	expirationTime := time.Hour
	

	accessToken, err := auth.MakeJWT(
		user.ID,
		cfg.Secret,
		expirationTime,
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create access JWT")
		return
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create refreshToken")
	}

	expireTime := time.Now().AddDate(0,0,60) 

   cfg.dbs.CreateTokenDB(r.Context(), database.CreateTokenDBParams{Token: refreshToken, UserID: user.ID,ExpiresAt: expireTime })


	respondWithJSON(w, http.StatusOK, response{
		User: User{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email:     user.Email,
		   Is_Chirpy_Red: user.IsChirpyRed.Bool,
		},
		Token: accessToken,
		RefreshToken: refreshToken,
	})


}

func (cfg *apiConfig) GetChirp(w http.ResponseWriter, r *http.Request) {

	currID := r.PathValue("chirpID")
	chirpID, err := uuid.Parse(currID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	chirp, err := cfg.dbs.GetChirpByID(r.Context(), chirpID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	chirpStruct := Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	}

	respondWithJSON(w, http.StatusOK, chirpStruct)

}

func (cfg *apiConfig) GetChirps(w http.ResponseWriter, r *http.Request) {

	dbResult, err := cfg.dbs.GetChirps(r.Context())
	if err != nil {
		msg := fmt.Sprintf("Error getting all chrips: %s", err)
		respondWithError(w, 500, msg)
		return

	}
	chirps := make([]Chirp, 0)
	for _, dbRow := range dbResult {
		chirps = append(chirps, Chirp{
			ID:        dbRow.ID,
			Body:      dbRow.Body,
			CreatedAt: dbRow.CreatedAt,
			UpdatedAt: dbRow.UpdatedAt,
			UserID:    dbRow.UserID,
		})
	}
	respondWithJSON(w, 200, chirps)
}

func (cfg *apiConfig) ResetDB(w http.ResponseWriter, r *http.Request) {

	if cfg.Platform != "dev" {
		msg := "Unauthorized Access"
		respondWithError(w, 403, msg)
		return
	} else {
		err := cfg.dbs.ResetTable(r.Context())
		if err != nil {
			msg := fmt.Sprintf("Error decoding parameters: %s", err)
			respondWithError(w, 500, msg)
			return
		}
		w.WriteHeader(200)
		w.Write([]byte("Hits reset to 0 and database reset to initial state."))
	}
}

func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request) {

	fmt.Fprintf(w, "Hits: %d", cfg.fileserverHits.Load())
}

func (cfg *apiConfig) CreateUser(w http.ResponseWriter, r *http.Request) {
	var userParams User
	decoder := json.NewDecoder(r.Body)

	err := decoder.Decode(&userParams)
	if err != nil {
		msg := fmt.Sprintf("Error decoding parameters: %s", err)
		respondWithError(w, 500, msg)
		return
	}

	hPass, err := auth.HashPassword(userParams.Password)
	if err != nil {
		fmt.Errorf("hashing password failed %s", err)
		return
	}

	dbUser, err := cfg.dbs.CreateUser(r.Context(), database.CreateUserParams{Email: userParams.Email, HashedPassword: hPass})
	if err != nil {
		msg := fmt.Sprintf("Error creating user for DB: %s", err)
		respondWithError(w, 500, msg)
		return
	}

	user := User{
		ID:        dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email:     dbUser.Email,
		Password:  dbUser.HashedPassword,
		Is_Chirpy_Red: dbUser.IsChirpyRed.Bool,

	}
	respondWithJSON(w, 201, user)

}

func FilterProfanity(in string) string {

	stringList := strings.Split(in, " ")

	for i := 0; i < len(stringList); i++ {

		if strings.ToLower(stringList[i]) == "kerfuffle" || strings.ToLower(stringList[i]) == "sharbert" || strings.ToLower(stringList[i]) == "fornax" {
			replacementString := "****"

			stringList[i] = replacementString
		}
	}
	return strings.Join(stringList, " ")

}

func respondWithError(w http.ResponseWriter, code int, msg string) {

	errResp := ErrorResponse{
		Error: msg}

	dat, err := json.Marshal(errResp)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(dat)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {

	dat, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(dat)

}
func (cfg *apiConfig) CreateChirp(w http.ResponseWriter, r *http.Request) {

type parameters struct {
		Body string `json:"body"`
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT")
		return
	}
	userID, err := auth.ValidateJWT(token, cfg.Secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT")
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	cleaned := FilterProfanity(params.Body)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	chirp, err := cfg.dbs.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   cleaned,
		UserID: userID,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create chirp")
		return
	}


	chirpStruct := Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	}

	respondWithJSON(w, 201, chirpStruct)

}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	secret := os.Getenv("SECRET")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Errorf("error opening database %s", err)
		return
	}

	dbQueries := database.New(db)

	mux := http.NewServeMux()
	var apiCfg apiConfig

	apiCfg.dbs = dbQueries
	apiCfg.Platform = platform
	apiCfg.Secret = secret

	handler := http.StripPrefix("/app/", http.FileServer(http.Dir('.')))

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(handler))
	servStruct := http.Server{
		Handler: mux,
		Addr:    ":8080",
	}
	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, req *http.Request) {

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	mux.HandleFunc("GET /admin/metrics", func(w http.ResponseWriter, req *http.Request) {

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		visitCount := apiCfg.fileserverHits.Load()
		fmt.Fprintf(w, "<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", visitCount)

	})

	mux.HandleFunc("POST /api/users", apiCfg.CreateUser)

	mux.HandleFunc("POST /admin/reset", apiCfg.ResetDB)

	mux.HandleFunc("POST /api/chirps", apiCfg.CreateChirp)

	mux.HandleFunc("GET /api/chirps", apiCfg.GetChirps)

	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.GetChirp)

	mux.HandleFunc("POST /api/login", apiCfg.Login)

	mux.HandleFunc("POST /api/refresh", apiCfg.Refresh)

	mux.HandleFunc("POST /api/revoke", apiCfg.Revoke)

	mux.HandleFunc("PUT /api/users", apiCfg.UpdateUserInfo)

	mux.HandleFunc("DELETE /api/chirps/{chirpID}", apiCfg.DeleteChirp)

	mux.HandleFunc("POST /api/polka/webhooks", apiCfg.UpgradeUser)

	err = servStruct.ListenAndServe()

	fmt.Print("err is: ", err)

}

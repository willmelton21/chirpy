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
}


type apiConfig struct {
	fileserverHits atomic.Int32
	dbs *database.Queries
   Platform string
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
	next.ServeHTTP(w,r)
	})
}

func (cfg *apiConfig) GetChirp(w http.ResponseWriter, r *http.Request) {
	
	currID := r.PathValue("chirpID")
	chirpID, err := uuid.Parse(currID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	
	chirp,err :=cfg.dbs.GetChirpByID(r.Context(),chirpID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

  chirpStruct := Chirp{
      ID: chirp.ID,
      CreatedAt: chirp.CreatedAt, 
      UpdatedAt: chirp.UpdatedAt,
      Body: chirp.Body, 
      UserID: chirp.UserID,

	
   }


	respondWithJSON(w,http.StatusOK,chirpStruct)

}

func (cfg *apiConfig) GetChirps(w http.ResponseWriter, r *http.Request) {
   
   dbResult,err := cfg.dbs.GetChirps(r.Context())
   if err != nil {
      msg := fmt.Sprintf("Error getting all chrips: %s",err)
      respondWithError(w,500,msg)
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
   respondWithJSON(w,200,chirps)
}

func (cfg *apiConfig) ResetDB(w http.ResponseWriter, r *http.Request) {
   
   if cfg.Platform != "dev" {
   msg := "Unauthorized Access"
     respondWithError(w,403,msg) 
     return
   } else {
      err := cfg.dbs.ResetTable(r.Context())  
      if err != nil {
	   msg := fmt.Sprintf("Error decoding parameters: %s",err)
         respondWithError(w,500,msg)
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
	   msg := fmt.Sprintf("Error decoding parameters: %s",err)
		respondWithError(w, 500, msg)
	  }	

	dbUser, err := cfg.dbs.CreateUser(r.Context(), userParams.Email)
	if err != nil {
	   msg := fmt.Sprintf("Error creating user for DB: %s",err)
		respondWithError(w, 500, msg)
		return
	 }
   
   user := User{
      ID: dbUser.ID,
      CreatedAt: dbUser.CreatedAt,
      UpdatedAt: dbUser.UpdatedAt,
      Email: dbUser.Email,
   }
	respondWithJSON(w,201,user)


 }

func FilterProfanity(in string) string{


	stringList := strings.Split(in," ")

	for i := 0; i < len(stringList); i++ {
		
		if strings.ToLower(stringList[i]) == "kerfuffle" || strings.ToLower(stringList[i]) == "sharbert" || strings.ToLower(stringList[i]) == "fornax" {
			replacementString := "****"
			
			stringList[i] = replacementString
		}
	}
	return strings.Join(stringList," ")

}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	
	errResp :=  ErrorResponse{
		Error: msg }

	dat, err := json.Marshal(errResp)
		if err != nil {
			log.Printf("Error marshalling JSON: %s",err)
      	w.WriteHeader(500)
			return
		  	} 
   w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(dat)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	
	dat ,err := json.Marshal(payload)
		if err != nil {
			log.Printf("Error marshalling JSON: %s",err)
      	w.WriteHeader(500)
			return
			}

   w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(dat)

	}
func (cfg *apiConfig) CreateChirp(w http.ResponseWriter, r *http.Request) {
   var chirpParams Chirp 
   decoder := json.NewDecoder(r.Body)
	
   err := decoder.Decode(&chirpParams)
   if err != nil {
	   msg := fmt.Sprintf("Error decoding parameters: %s",err)
		respondWithError(w, 500, msg)
	  }	

  	if len(chirpParams.Body) > 140 {
		respondWithError(w,400,"Chirp is too long")
      return
	}

	chirpParams.Body = FilterProfanity(chirpParams.Body)

   CreatedChirp, err := cfg.dbs.CreateChirp(r.Context(), database.CreateChirpParams{Body: chirpParams.Body,UserID: chirpParams.UserID})
	if err != nil {
      msg := fmt.Sprintf("Error creating chirp for user: %s",err)
		respondWithError(w, 500, msg)
		return
	 }

   chirp := Chirp{
      ID: CreatedChirp.ID,
      CreatedAt: CreatedChirp.CreatedAt, 
      UpdatedAt: CreatedChirp.UpdatedAt,
      Body: CreatedChirp.Body, 
      UserID: CreatedChirp.UserID,
   }


	respondWithJSON(w,201,chirp)
	
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
   platform := os.Getenv("PLATFORM")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Errorf("error opening database %s",err)
		return
	}

	dbQueries := database.New(db)


	mux := http.NewServeMux()
	var apiCfg apiConfig

	apiCfg.dbs = dbQueries
   apiCfg.Platform = platform
	
		

	handler := http.StripPrefix("/app/",http.FileServer(http.Dir('.')))


	mux.Handle("/app/",apiCfg.middlewareMetricsInc(handler))
	servStruct := http.Server{
		Handler: mux,
		Addr: ":8080",
	}
	mux.HandleFunc("GET /api/healthz",func(w http.ResponseWriter, req *http.Request) {
		
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	mux.HandleFunc("GET /admin/metrics",func(w http.ResponseWriter, req *http.Request) {

      w.Header().Set("Content-Type", "text/html; charset=utf-8")
      w.WriteHeader(http.StatusOK)
      visitCount := apiCfg.fileserverHits.Load()
      fmt.Fprintf(w, "<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>",visitCount)
      
      })


	mux.HandleFunc("POST /api/users",apiCfg.CreateUser)

   mux.HandleFunc("POST /admin/reset",apiCfg.ResetDB)

   mux.HandleFunc("POST /api/chirps",apiCfg.CreateChirp)

   mux.HandleFunc("GET /api/chirps",apiCfg.GetChirps)
	
   mux.HandleFunc("GET /api/chirps/{chirpID}",apiCfg.GetChirp)

	err = servStruct.ListenAndServe()

	fmt.Print("err is: ",err)

}

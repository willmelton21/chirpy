package main

import (
	"net/http"
	"fmt"
   "sync/atomic"	
   "encoding/json"
   "log"
)


type apiConfig struct {
	fileserverHits atomic.Int32
}


func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w,r)
	})
}

func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request) {

	fmt.Fprintf(w, "Hits: %d", cfg.fileserverHits.Load())
}
func (cfg *apiConfig) resetHandler(w http.ResponseWriter, r *http.Request) {

	cfg.fileserverHits.Store(0)
}

func (cfg *apiConfig) validate_chirp(w http.ResponseWriter, r *http.Request) {
   
   type parameters struct {
      body string `json:"body"`
      }

   decoder := json.NewDecoder(r.Body)
   params := parameters{}
   err := decoder.Decode(&params)
   if err != nil {
      
      log.Printf("Error decoding parameters: %s",err)
      w.WriteHeader(500)
      return
   }



   dat, err := json.Marshal(params)
   if err != nil {
      log.Printf("Error marshalling JSON: %s",err)
      w.WriteHeader(500)
      return
   }

   w.Header().Set("Content-Type", "application/json")
   w.WriteHeader(200)
   w.Write(dat)
}

func main() {

	mux := http.NewServeMux()
	var apiCfg apiConfig

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
	mux.HandleFunc("POST /admin/reset",apiCfg.resetHandler)

   mux.HandleFunc("POST /api/validate_chirp",apiCfg.validate_chirp)
	
	err := servStruct.ListenAndServe()

	fmt.Print("err is: ",err)

}

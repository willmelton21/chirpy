package main

import (
	"net/http"
	"fmt"
   "sync/atomic"	
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
		
		w.Header().Set("Content_Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	mux.HandleFunc("GET /api/metrics",apiCfg.metricsHandler)
	mux.HandleFunc("POST /api/reset",apiCfg.resetHandler)
	
	err := servStruct.ListenAndServe()

	fmt.Print("err is: ",err)

}

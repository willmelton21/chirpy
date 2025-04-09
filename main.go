package main

import (
	"net/http"
	"fmt"
)

func main() {

	mux := http.NewServeMux()

	mux.Handle("/app/",http.StripPrefix("/app/",http.FileServer(http.Dir('.'))))
	servStruct := http.Server{
		Handler: mux,
		Addr: ":8080",
	}
	mux.HandleFunc("/healthz",func(w http.ResponseWriter, req *http.Request) {
		
		w.Header().Set("Content_Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	err := servStruct.ListenAndServe()

	fmt.Print("err is: ",err)

}

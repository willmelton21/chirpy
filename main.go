package main

import (
	"net/http"
	"fmt"
)

func main() {

	mux := http.NewServeMux()

	servStruct := http.Server{
		Handler: mux,
		Addr: ":8080",
	}
	err := servStruct.ListenAndServe()

	fmt.Print("err is: ",err)

}

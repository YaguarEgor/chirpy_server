package main

import (
	"net/http"
)

func main() {
	const filepathRoot = "."
	const port = "8080"

	//serverMux := http.ServeMux{}
	mux := http.NewServeMux()
	mux.Handle("/app/", http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot))))
	mux.HandleFunc("/healthz", handlerReadiness)
	server := &http.Server{
		Handler: mux,
		Addr:    ":"+port,
	}
	server.ListenAndServe()

}

func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}
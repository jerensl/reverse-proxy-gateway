package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

func main() {
	handler := http.HandlerFunc(UsersReverseProxy)

	fmt.Printf("Starting users service at port: %v", os.Getenv("PORT"))
	if err := http.ListenAndServe(":"+os.Getenv("PORT"), handler); err != nil {
		panic(err)
	}
}

func UsersReverseProxy(w http.ResponseWriter, r *http.Request) {
	host, err := url.Parse(os.Getenv("USERS_SERVICE"))
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		return
	}

	reverseProxy := httputil.NewSingleHostReverseProxy(host)

	reverseProxy.ServeHTTP(w, r)
}
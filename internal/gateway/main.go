package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"sync/atomic"
)

func main() {
	users := strings.Split(os.Getenv("USERS_SERVICE"), ";")

	for _, user := range users {
		serviceUrl, err := url.Parse(user)
		if err != nil {
			log.Fatal(err)
		}

		reverseProxy := httputil.NewSingleHostReverseProxy(serviceUrl)

		serverPool.AddServer(&Backend{
			URL: serviceUrl,
			ReverseProxy: reverseProxy,
		})
	} 

	handler := http.HandlerFunc(UsersLoadBalancer)

	fmt.Printf("Starting users service at port: %v", os.Getenv("PORT"))
	if err := http.ListenAndServe(":"+os.Getenv("PORT"), handler); err != nil {
		panic(err)
	}
}

type Backend struct {
	URL				*url.URL
	ReverseProxy 	*httputil.ReverseProxy
}

type ServerPool struct {
	Backends []*Backend
	current uint64
}

func (bp *ServerPool) AddServer(backend *Backend) {
	bp.Backends = append(bp.Backends, backend)
}

func (bp *ServerPool) GetNextIndex() int {
	return int(atomic.AddUint64(&bp.current, uint64(1)) % uint64(len(bp.Backends)))
}

func (bp *ServerPool) GetNextServer() *Backend {
	start := bp.GetNextIndex()
	end := len(bp.Backends) + start

	for i := start; i < end; i++ {
		index := i % len(bp.Backends)

		atomic.StoreUint64(&bp.current, uint64(index))

		return bp.Backends[index]
	}

	return nil
}

var serverPool ServerPool

func UsersLoadBalancer(w http.ResponseWriter, r *http.Request) {
	server := serverPool.GetNextServer()

	server.ReverseProxy.ServeHTTP(w, r)
}

func UsersReverseProxy(w http.ResponseWriter, r *http.Request) {
	host, err := url.Parse(os.Getenv("USERS_SERVICE"))
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		return
	}

	reverseProxy := httputil.NewSingleHostReverseProxy(host)
	if reverseProxy != nil {
		reverseProxy.ServeHTTP(w, r)
	}
	http.Error(w, "Service not available", http.StatusServiceUnavailable)
}
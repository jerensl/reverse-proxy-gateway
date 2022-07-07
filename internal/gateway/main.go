package main

import (
	"context"
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
		serverPool.AddServer(user)
	} 

	http.HandleFunc("/", UsersLoadBalancer)

	fmt.Printf("Starting users service at port: %v", os.Getenv("PORT"))
	if err := http.ListenAndServe(":"+os.Getenv("PORT"), nil); err != nil {
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

func (bp *ServerPool) AddServer(host string) {
	serviceUrl, err := url.Parse(host)
	if err != nil {
		log.Fatal(err)
	}
	
	reverseProxy := httputil.NewSingleHostReverseProxy(serviceUrl)
	reverseProxy.ErrorHandler = UsersLoadBalancerErrorHandler

	bp.Backends = append(bp.Backends, &Backend{
		URL: serviceUrl,
		ReverseProxy: reverseProxy,
	})
}

func (bp *ServerPool) GetNextIndex() int {
	if int(atomic.LoadUint64(&bp.current)) > len(bp.Backends)*2-1 {
		atomic.StoreUint64(&bp.current, 0)
	}

	return int(atomic.AddUint64(&bp.current, uint64(1)) % uint64(len(bp.Backends)))
}

func (bp *ServerPool) GetNextServer() *Backend {
	index := bp.GetNextIndex()

	return bp.Backends[index]
}

var serverPool ServerPool

const Visit int = 1

func GetVisitingNodeFromContext(r *http.Request) int {
	if visit, ok := r.Context().Value(Visit).(int); ok {
		return visit
	}

	return 0
}

func UsersLoadBalancer(w http.ResponseWriter, r *http.Request) {
	server := serverPool.GetNextServer()

	server.ReverseProxy.ServeHTTP(w, r)
}

func UsersLoadBalancerErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	visiting := GetVisitingNodeFromContext(r)

	if visiting > len(serverPool.Backends) {
		
		http.Error(w, "Service not available", http.StatusServiceUnavailable)
		return
	}
	
	ctx := context.WithValue(r.Context(), Visit, visiting+1)
	UsersLoadBalancer(w, r.WithContext(ctx))
}
package social

import (
	"appengine"
	"fmt"
	"net/http"
	"os"
	"strings"
)

func InitServiceRoutes() {
	http.HandleFunc("/health/", healthHandler)
	http.HandleFunc("/users/", userHandler)
	http.HandleFunc("/tasks/daily", dailyTaskHandler)
	http.HandleFunc("/tasks/worker", workerTaskHandler)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "All your base are belong to us!")
}

func userHandler(w http.ResponseWriter, r *http.Request) {
	if !checkAuthorization(r) {
		http.Error(w, "👹", http.StatusUnauthorized)
		return
	}

	switch r.Method {
	case "GET":
		if err := GetUser(w, r); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	case "POST":
		if err := CreateUser(w, r); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	default:
		http.Error(w, "☠️", http.StatusMethodNotAllowed)
	}
}

func dailyTaskHandler(w http.ResponseWriter, r *http.Request) {
	err := CreateRequestQueue(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
}

func workerTaskHandler(_ http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	CreateWorker(&ctx)
}

func checkAuthorization(r *http.Request) bool {
	key, exist := os.LookupEnv("FORD_AUTH_TOKEN")

	if !exist {
		return true
	}

	auth := r.Header.Get("Authorization")

	return strings.Compare(auth, key) == 0
}

func setResponseHeader(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
}

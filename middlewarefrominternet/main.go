// You can edit this code!
// Click here and start typing.
package main

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type ctxKey uint8

const userKey ctxKey = 0

type user struct {
	name string
}

func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := &user{}
		r = r.WithContext(context.WithValue(r.Context(), userKey, u))

		defer func(start time.Time) {
			if u, ok := r.Context().Value(userKey).(*user); ok {
				fmt.Fprintf(w, "log hello %s: %s\n", u.name, time.Now())
			} else {
				fmt.Fprintf(w, "log hello\n")
			}
		}(time.Now())

		next.ServeHTTP(w, r)
	})
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if u, ok := r.Context().Value(userKey).(*user); ok {
			u.name = "user123"
			fmt.Println("authorized")
		} else {
			u := &user{name: "anonymous"}
			r = r.WithContext(context.WithValue(r.Context(), userKey, u))
		}
		next.ServeHTTP(w, r)
	})
}

func welcome(w http.ResponseWriter, r *http.Request) {
	fmt.Println("welcome")
	time.Sleep(time.Second * 5) // arbitrary sleep to see that logMiddleware does its job as expected
	if u, ok := r.Context().Value(userKey).(*user); ok {
		fmt.Fprintf(w, "hello %s\n", u.name)
	} else {
		fmt.Fprintf(w, "hello\n")
	}
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/welcome", welcome)
	chain := logMiddleware(authMiddleware(mux))
	http.ListenAndServe("localhost:9090", chain)

}

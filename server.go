package main

import (
 "encoding/json"
 "fmt"
 "math/rand"
 "net/http"
 "time"
)

func main() {
 http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
  w.Write([]byte("Boost Server Live OK"))
 })
 
 http.HandleFunc("/create", func(w http.ResponseWriter, r *http.Request) {
  rand.Seed(time.Now().UnixNano())
  key := fmt.Sprintf("NG-%d", rand.Intn(999999))
  json.NewEncoder(w).Encode(map[string]string{"key": key})
 })
 
 http.HandleFunc("/verify", func(w http.ResponseWriter, r *http.Request) {
  json.NewEncoder(w).Encode(map[string]interface{}{"valid": true, "days_left": 30})
 })

 http.ListenAndServe(":10000", nil)
}

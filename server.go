package main

import (
 "encoding/json"
 "fmt"
 "io"
 "math/rand"
 "net/http"
 "os"
 "sync"
 "time"
)

var (
 mu sync.Mutex
 keys = map[string]string{}
 refs = map[string]string{}
)

func verifyPaystack(ref string, secret string) bool {
 if ref == "" { return false }
 req, _ := http.NewRequest("GET", "https://api.paystack.co/transaction/verify/"+ref, nil)
 req.Header.Set("Authorization", "Bearer "+secret)
 client := &http.Client{Timeout: 10 * time.Second}
 resp, err := client.Do(req)
 if err != nil { return false }
 defer resp.Body.Close()
 body, _ := io.ReadAll(resp.Body)
 var data struct {
  Status bool `json:"status"`
  Data struct {
   Status string `json:"status"`
   Amount int `json:"amount"`
  } `json:"data"`
 }
 if err := json.Unmarshal(body, &data); err != nil { return false }
 return data.Status && data.Data.Status == "success" && data.Data.Amount >= 145000
}

func main() {
 secret := os.Getenv("PAYSTACK_SECRET")
 http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
  w.Write([]byte("Boost Server Live OK - SECURE"))
 })
 http.HandleFunc("/create", func(w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Access-Control-Allow-Origin", "*")
  ref := r.URL.Query().Get("ref")
  if ref == "" { http.Error(w, "no ref", 400); return }
  mu.Lock()
  if existing, ok := refs[ref]; ok {
   mu.Unlock()
   json.NewEncoder(w).Encode(map[string]string{"key": existing})
   return
  }
  mu.Unlock()
  if secret != "" {
   if !verifyPaystack(ref, secret) {
    w.WriteHeader(403)
    json.NewEncoder(w).Encode(map[string]string{"error": "payment_not_verified"})
    return
   }
  }
  rand.Seed(time.Now().UnixNano())
  mu.Lock()
  k := fmt.Sprintf("NG-%06d", rand.Intn(900000)+100000)
  keys[k] = ""
  refs[ref] = k
  mu.Unlock()
  json.NewEncoder(w).Encode(map[string]string{"key": k})
 })
 http.HandleFunc("/verify", func(w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Access-Control-Allow-Origin", "*")
  key := r.URL.Query().Get("key")
  hwid := r.URL.Query().Get("hwid")
  mu.Lock()
  defer mu.Unlock()
  storedHwid, exists := keys[key]
  if !exists { json.NewEncoder(w).Encode(map[string]interface{}{"valid": false}); return }
  if storedHwid == "" {
   keys[key] = hwid
   json.NewEncoder(w).Encode(map[string]interface{}{"valid": true, "days_left": 30})
   return
  }
  if storedHwid == hwid {
   json.NewEncoder(w).Encode(map[string]interface{}{"valid": true, "days_left": 30})
   return
  }
  json.NewEncoder(w).Encode(map[string]interface{}{"valid": false, "reason": "already used"})
 })
 port := os.Getenv("PORT")
 if port == "" { port = "10000" }
 http.ListenAndServe(":"+port, nil)
}

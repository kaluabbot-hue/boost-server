package main

import (
 "crypto/sha256"
 "encoding/json"
 "fmt"
 "math/rand"
 "net/http"
 "sync"
 "time"
)

type License struct {
 Key string
 ExpiresAt time.Time
 HWID string
}

var (
 licenses = make(map[string]License)
 mu sync.RWMutex
)

func generateKey() string {
 rand.Seed(time.Now().UnixNano())
 chars := "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
 b := make([]byte, 12)
 for i := range b {
  b[i] = chars[rand.Intn(len(chars))]
 }
 return fmt.Sprintf("NG-%s-%s-%s", string(b[0:4]), string(b[4:8]), string(b[8:12]))
}

func createHandler(w http.ResponseWriter, r *http.Request) {
 days := 30
 mu.Lock()
 key := generateKey()
 licenses[key] = License{Key: key, ExpiresAt: time.Now().AddDate(0,0,days)}
 mu.Unlock()
 json.NewEncoder(w).Encode(map[string]interface{}{"key": key, "days": days})
}

func verifyHandler(w http.ResponseWriter, r *http.Request) {
 key := r.URL.Query().Get("key")
 hwid := r.URL.Query().Get("hwid")
 mu.Lock()
 defer mu.Unlock()
 lic, ok := licenses[key]
 if !ok {
  json.NewEncoder(w).Encode(map[string]bool{"valid": false})
  return
 }
 if time.Now().After(lic.ExpiresAt) {
  delete(licenses, key)
  json.NewEncoder(w).Encode(map[string]bool{"valid": false})
  return
 }
 if lic.HWID == "" && hwid != "" {
  lic.HWID = hwid
  licenses[key] = lic
 }
 if lic.HWID != "" && lic.HWID != hwid {
  json.NewEncoder(w).Encode(map[string]interface{}{"valid": false, "reason": "hwid_mismatch"})
  return
 }
 daysLeft := int(time.Until(lic.ExpiresAt).Hours() / 24)
 json.NewEncoder(w).Encode(map[string]interface{}{"valid": true, "days_left": daysLeft})
}

func main() {
 http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
  w.Write([]byte("Boost Server Live - 10/10 HWID"))
 })
 http.HandleFunc("/create", createHandler)
 http.HandleFunc("/verify", verifyHandler)
 http.ListenAndServe(":10000", nil)
}

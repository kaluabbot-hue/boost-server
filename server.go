// FINAL SERVER.GO - DO NOT CHANGE AGAIN
package main

import (
 "encoding/json"
 "io"
 "net/http"
 "os"
 "time"
)

func verifyPaystack(ref string, secret string) bool {
 if ref == "" { return false }
 req, _ := http.NewRequest("GET", "https://api.paystack.co/transaction/verify/"+ref, nil)
 req.Header.Set("Authorization", "Bearer "+secret)
 client := &http.Client{Timeout: 10*time.Second}
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
 json.Unmarshal(body, &data)
 return data.Status && data.Data.Status == "success" && data.Data.Amount >= 145000
}

func main() {
 secret := os.Getenv("PAYSTACK_SECRET")
 http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("Boost Server Live")) })
 http.HandleFunc("/stabilizer.sh", func(w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Content-Type", "text/plain")
  w.Write([]byte(os.Getenv("STABILIZER_CODE")))
 })
 http.HandleFunc("/license", func(w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Content-Type", "text/plain")
  ref := r.URL.Query().Get("ref")
  if ref == "" { http.Error(w, "no ref", 400); return }
  if secret != "" { if !verifyPaystack(ref, secret) { w.WriteHeader(403); w.Write([]byte("payment_not_verified")); return } }
  w.Write([]byte(ref))
 })
 port := os.Getenv("PORT")
 if port == "" { port = "10000" }
 http.ListenAndServe(":"+port, nil)
}

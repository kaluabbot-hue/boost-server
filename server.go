package main

import (
 "encoding/json"
 "io"
 "net/http"
 "os"
 "time"
)

func verifyPaystack(ref string, secret string) (bool, time.Time) {
 if ref == "" { return false, time.Time{} }
 req, _ := http.NewRequest("GET", "https://api.paystack.co/transaction/verify/"+ref, nil)
 req.Header.Set("Authorization", "Bearer "+secret)
 client := &http.Client{Timeout: 10*time.Second}
 resp, err := client.Do(req)
 if err != nil { return false, time.Time{} }
 defer resp.Body.Close()
 body, _ := io.ReadAll(resp.Body)
 var data struct {
  Status bool `json:"status"`
  Data struct {
   Status string `json:"status"`
   Amount int `json:"amount"`
   PaidAt string `json:"paid_at"`
  } `json:"data"`
 }
 json.Unmarshal(body, &data)
 if !(data.Status && data.Data.Status == "success" && data.Data.Amount >= 145000) {
  return false, time.Time{}
 }
 t, _ := time.Parse(time.RFC3339, data.Data.PaidAt)
 if t.IsZero() { t = time.Now() }
 return true, t
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
  ok, paidAt := verifyPaystack(ref, secret)
  if secret != "" && !ok { w.WriteHeader(403); w.Write([]byte("payment_not_verified")); return }
  if paidAt.IsZero() { paidAt = time.Now() }
  exp := paidAt.AddDate(0, 0, 30)
  w.Write([]byte("REF="+ref+"\nEXP="+exp.Format("2006-01-02")+"\n"))
 })
 port := os.Getenv("PORT")
 if port == "" { port = "10000" }
 http.ListenAndServe(":"+port, nil)
}

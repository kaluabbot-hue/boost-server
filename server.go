package main

import (
    "encoding/json"
    "fmt"
    "math/rand"
    "net/http"
    "os"
    "strings"
    "time"
)

func randomString(n int) string {
    letters := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
    b := make([]rune, n)
    for i := range b {
        b[i] = letters[rand.Intn(len(letters))]
    }
    return string(b)
}

func verifyHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Access-Control-Allow-Origin", "*")
    ref := r.URL.Query().Get("reference")
    if ref == "" {
        json.NewEncoder(w).Encode(map[string]string{"error": "No reference"})
        return
    }
    // Read secret key from Render Environment - you just saved it!
    secret := os.Getenv("PAYSTACK_SECRET_KEY")
    fmt.Println("Using secret key:", secret[:10]+"...")

    // Generate 30-day license
    license := "NS-30D-" + strings.ToUpper(randomString(12))
    expiry := time.Now().AddDate(0, 1, 0)

    json.NewEncoder(w).Encode(map[string]interface{}{
        "license":   license,
        "expiry":    expiry.Format("2006-01-02"),
        "reference": ref,
    })
}

func main() {
    http.HandleFunc("/api/verify", verifyHandler)
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Server Live! License API Ready!"))
    })
    port := os.Getenv("PORT")
    if port == "" {
        port = "10000"
    }
    fmt.Println("Server starting on port", port)
    http.ListenAndServe(":"+port, nil)
}

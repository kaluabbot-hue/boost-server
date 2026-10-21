package main

import (
    "encoding/json"
    "fmt"
    "io"
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
    w.Header().Set("Content-Type", "application/json")
    
    ref := r.URL.Query().Get("reference")
    if ref == "" {
        json.NewEncoder(w).Encode(map[string]interface{}{"status": "error", "message": "No reference"})
        return
    }

    // REAL PAYSTACK VERIFICATION
    secret := os.Getenv("PAYSTACK_SECRET_KEY")
    if secret == "" {
        json.NewEncoder(w).Encode(map[string]interface{}{"status": "error", "message": "Server key not set"})
        return
    }

    // Call Paystack to verify this reference is real and paid!
    req, _ := http.NewRequest("GET", "https://api.paystack.co/transaction/verify/"+ref, nil)
    req.Header.Set("Authorization", "Bearer "+secret)
    client := &http.Client{Timeout: 10 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        json.NewEncoder(w).Encode(map[string]interface{}{"status": "error", "message": "Paystack error"})
        return
    }
    defer resp.Body.Close()
    body, _ := io.ReadAll(resp.Body)

    var result map[string]interface{}
    json.Unmarshal(body, &result)

    // Check if Paystack says payment is SUCCESS
    if status, ok := result["status"].(bool); !ok || !status {
        json.NewEncoder(w).Encode(map[string]interface{}{"status": "error", "message": "Paystack verification failed"})
        return
    }
    data, ok := result["data"].(map[string]interface{})
    if !ok || data["status"] != "success" {
        json.NewEncoder(w).Encode(map[string]interface{}{"status": "error", "message": "Payment not successful"})
        return
    }

    // REAL PAYMENT CONFIRMED! NOW GENERATE 30-DAY LICENSE!
    license := "NS-30D-" + strings.ToUpper(randomString(12))
    expiry := time.Now().AddDate(0, 1, 0)

    json.NewEncoder(w).Encode(map[string]interface{}{
        "status":    "success",
        "license":   license,
        "expiry":    expiry.Format("2006-01-02"),
        "reference": ref,
        "amount":    data["amount"],
        "email":     data["customer"].(map[string]interface{})["email"],
    })
}

func main() {
    http.HandleFunc("/api/verify", verifyHandler)
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Real Paystack Verification Server Live!"))
    })
    port := os.Getenv("PORT")
    if port == "" {
        port = "10000"
    }
    fmt.Println("Real server starting on port", port)
    http.ListenAndServe(":"+port, nil)
}

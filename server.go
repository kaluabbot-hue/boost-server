package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type License struct {
	Key       string    `json:"key"`
	HWID      string    `json:"hwid"`
	Expiry    time.Time `json:"expiry"`
	CreatedAt time.Time `json:"created_at"`
	Email     string    `json:"email"`
}

var (
	licenses = make(map[string]License)
	mu       sync.RWMutex
)

func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func verifyHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(200)
		return
	}

	key := r.URL.Query().Get("key")
	hwid := r.URL.Query().Get("hwid")

	if key == "" || hwid == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"valid": false,
			"message": "key and hwid required",
		})
		return
	}

	mu.RLock()
	lic, exists := licenses[key]
	mu.RUnlock()

	// FOR LAUNCH: If key doesn't exist, auto-create for testing (remove this after you add Paystack)
	if !exists {
		// Only allow keys that start with NG- for security
		if len(key) < 8 || key[:3] != "NG-" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"valid": false,
				"message": "invalid key format",
			})
			return
		}
		// Create 30 days license for this key
		newLic := License{
			Key:       key,
			HWID:      hwid,
			Expiry:    time.Now().Add(30 * 24 * time.Hour),
			CreatedAt: time.Now(),
		}
		mu.Lock()
		licenses[key] = newLic
		lic = newLic
		exists = true
		mu.Unlock()
	}

	// Check HWID binding - 1 key = 1 device
	if lic.HWID != "" && lic.HWID != hwid {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"valid": false,
			"message": "key already used on another device",
		})
		return
	}

	// Bind HWID on first use
	if lic.HWID == "" || lic.HWID != hwid {
		mu.Lock()
		lic.HWID = hwid
		licenses[key] = lic
		mu.Unlock()
	}

	serverTime := time.Now().UTC()
	isExpired := serverTime.After(lic.Expiry)
	daysLeft := int(lic.Expiry.Sub(serverTime).Hours() / 24)
	if daysLeft < 0 {
		daysLeft = 0
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"valid":      !isExpired && exists,
		"expiry":     lic.Expiry.Format(time.RFC3339),
		"server_time": serverTime.Format(time.RFC3339),
		"days_left":  daysLeft,
		"message":    "ok",
	})
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "online",
		"time": time.Now().UTC().Format(time.RFC3339),
	})
}

func main() {
	http.HandleFunc("/verify", verifyHandler)
	http.HandleFunc("/", healthHandler)
	http.HandleFunc("/health", healthHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Nigeria Boost Server running on port %s\n", port)
	fmt.Printf("Verify URL: /verify?key=NG-XXXX&hwid=YOUR_HWID\n")
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

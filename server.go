package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"sync"
)

var (
	boosts = make(map[string]time.Time)
	mu sync.Mutex
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Boost Server Live - 30 Days")
	})

	// Check if user is on cooldown
	http.HandleFunc("/check", func(w http.ResponseWriter, r *http.Request) {
		userID := r.URL.Query().Get("user_id")
		mu.Lock()
		expiry, exists := boosts[userID]
		mu.Unlock()

		if exists && time.Now().Before(expiry) {
			daysLeft := int(expiry.Sub(time.Now()).Hours() / 24)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"on_cooldown": true,
				"days_left": daysLeft,
			})
		} else {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"on_cooldown": false,
			})
		}
	})

	// Verify payment and start boost
	http.HandleFunc("/verify", func(w http.ResponseWriter, r *http.Request) {
		userID := r.URL.Query().Get("user_id")
		mu.Lock()
		boosts[userID] = time.Now().AddDate(0, 0, 30) // 30 DAYS
		mu.Unlock()
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "boost_started",
			"expires": boosts[userID],
		})
	})

	http.ListenAndServe(":10000", nil)
}

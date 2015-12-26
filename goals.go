package goals

import (
	"fmt"
	"net/http"
)

func init() {
	http.HandleFunc("/", handler)
	http.HandleFunc("/fitbit/auth", FitbitAuthHandler)
	http.HandleFunc("/fitbit/callback", FitbitAuthCallbackHandler)
	http.HandleFunc("/fitbit/profile", GetProfile)
	http.HandleFunc("/admin/settings", SettingsHandler)
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "This should be my goals app!")
}

package goals

import (
	"encoding/json"
	"fmt"
	"net/http"

	"google.golang.org/appengine"
)

func init() {
	http.HandleFunc("/", handler)
	http.HandleFunc("/fitbit/auth", FitbitAuthHandler)
	http.HandleFunc("/fitbit/callback", FitbitAuthCallbackHandler)
	http.HandleFunc("/admin/settings", SettingsHandler)

	http.HandleFunc("/api/goals", goalsHandler)
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "This should be my goals app!")
}

type GoalsResponse struct {
	Steps DailyStepGoals `json:"steps"`
}

func goalsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	stepGoals, _ := getDailyStepGoals(ctx)

	response := &GoalsResponse{
		Steps: stepGoals,
	}

	b, _ := json.Marshal(response)

	fmt.Fprintf(w, "%s", b)
}

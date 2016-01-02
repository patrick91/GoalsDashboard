package goals

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/memcache"
)

// Config is a custom oauth2 config that is used to store the token
// in the datastore
type Config struct {
	*oauth2.Config
}

// StoreToken is called when exchanging the token and saves the token
// in the datastore
func (c *Config) StoreToken(ctx context.Context, token *oauth2.Token) error {
	log.Infof(ctx, "storing the token")

	key := datastore.NewKey(ctx, "Tokens", "fitbit", 0, nil)

	_, err := datastore.Put(ctx, key, token)

	return err
}

// Exchange is a wrapper around oauth2.config.Exchange and stores the Token
// in the datastore
func (c *Config) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	token, err := c.Config.Exchange(ctx, code)

	if err != nil {
		return nil, err
	}

	if err := c.StoreToken(ctx, token); err != nil {
		return nil, err
	}

	return token, nil
}

// Client creates a new client using our custom TokenSource
func (c *Config) Client(ctx context.Context, t *oauth2.Token) *http.Client {
	return oauth2.NewClient(ctx, c.TokenSource(ctx, t))
}

// TokenSource uses uses our DatastoreTokenSource as the source token
func (c *Config) TokenSource(ctx context.Context, t *oauth2.Token) oauth2.TokenSource {
	rts := &DatastoreTokenSource{
		source: c.Config.TokenSource(ctx, t),
		config: c,
		ctx:    ctx,
	}

	return oauth2.ReuseTokenSource(t, rts)
}

// DatastoreTokenSource is our custom TokenSource
type DatastoreTokenSource struct {
	config *Config
	source oauth2.TokenSource
	ctx    context.Context
}

// Token saves the token in the datastore when it is updated
func (t *DatastoreTokenSource) Token() (*oauth2.Token, error) {
	token, err := t.source.Token()

	if err != nil {
		return nil, err
	}

	if err := t.config.StoreToken(t.ctx, token); err != nil {
		return nil, err
	}

	return token, nil
}

func getFitbitConf(ctx context.Context) (*Config, error) {
	var data Settings

	err := GetSettings(ctx, &data)

	if err != nil {
		return nil, err
	}

	var fitbitConf = &Config{
		Config: &oauth2.Config{
			ClientID:     data.FitbitClientID,
			ClientSecret: data.FitbitClientSecret,
			Scopes:       []string{"activity", "weight", "profile"},
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://www.fitbit.com/oauth2/authorize",
				TokenURL: "https://api.fitbit.com/oauth2/token",
			},
		},
	}

	return fitbitConf, nil
}

// FitbitAuthHandler authenticates the user against FitBit
func FitbitAuthHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	fitbitConf, err := getFitbitConf(ctx)

	if err != nil {
		fmt.Fprint(w, "Remember to initialise your settings")

		return
	}

	// Redirect user to consent page to ask for permission
	// for the scopes specified above.
	url := fitbitConf.AuthCodeURL("state", oauth2.AccessTypeOffline)

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func getToken(ctx context.Context, data *oauth2.Token) error {
	key := datastore.NewKey(ctx, "Tokens", "fitbit", 0, nil)

	err := datastore.Get(ctx, key, data)

	return err
}

// FitbitAuthCallbackHandler stores the token received from FitBit
func FitbitAuthCallbackHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	fitbitConf, err := getFitbitConf(ctx)

	if err != nil {
		fmt.Fprint(w, "Remember to initialise your settings")

		return
	}

	code := r.URL.Query()["code"][0]

	// Use the authorization code that is pushed to the redirect URL.
	// NewTransportWithCode will do the handshake to retrieve
	// an access token and initiate a Transport that is
	// authorized and authenticated by the retrieved token.
	_, err = fitbitConf.Exchange(ctx, code)

	if err != nil {
		log.Errorf(ctx, "%v", err)
	}

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func getFitbitClient(ctx context.Context) (*http.Client, error) {
	var token oauth2.Token

	fitbitConf, err := getFitbitConf(ctx)

	if err != nil {
		return nil, err
	}

	err = getToken(ctx, &token)

	if err != nil {
		return nil, err
	}

	client := fitbitConf.Client(ctx, &token)

	return client, nil
}

type ActivitiesOutput struct {
	Activities []interface{} `json:"activities"`
	Goals      struct {
		ActiveMinutes int     `json:"activeMinutes"`
		CaloriesOut   int     `json:"caloriesOut"`
		Distance      float64 `json:"distance"`
		Floors        int     `json:"floors"`
		Steps         int     `json:"steps"`
	} `json:"goals"`
	Summary struct {
		ActiveScore      int `json:"activeScore"`
		ActivityCalories int `json:"activityCalories"`
		CaloriesBMR      int `json:"caloriesBMR"`
		CaloriesOut      int `json:"caloriesOut"`
		Distances        []struct {
			Activity string  `json:"activity"`
			Distance float64 `json:"distance"`
		} `json:"distances"`
		Elevation            float64 `json:"elevation"`
		FairlyActiveMinutes  int     `json:"fairlyActiveMinutes"`
		Floors               int     `json:"floors"`
		LightlyActiveMinutes int     `json:"lightlyActiveMinutes"`
		MarginalCalories     int     `json:"marginalCalories"`
		SedentaryMinutes     int     `json:"sedentaryMinutes"`
		Steps                int     `json:"steps"`
		VeryActiveMinutes    int     `json:"veryActiveMinutes"`
	} `json:"summary"`
}

var key = "dailyGoals"

type DailyStepGoals struct {
	Current int `json:"current"`
	Goal    int `json:"goal"`
}

func storeDailyGoals(ctx context.Context, goals DailyStepGoals) {
	duration, _ := time.ParseDuration("10s")

	item := &memcache.Item{
		Key:        key,
		Object:     goals,
		Expiration: duration,
	}

	memcache.Gob.Set(ctx, item)
}

func getDailyStepGoals(ctx context.Context) (DailyStepGoals, error) {
	goals := DailyStepGoals{}

	if _, err := memcache.Gob.Get(ctx, key, &goals); err == memcache.ErrCacheMiss {
		t := time.Now()

		url := fmt.Sprintf(
			"https://api.fitbit.com/1/user/-/activities/date/%d-%d-%d.json",
			t.Year(), t.Month(), t.Day(),
		)

		client, err := getFitbitClient(ctx)

		if err != nil {
			return goals, err
		}

		res, err := client.Get(url)

		if err != nil {
			log.Errorf(ctx, "%v", err)
		}

		output := new(ActivitiesOutput)

		defer res.Body.Close()

		err = json.NewDecoder(res.Body).Decode(output)

		goals.Current = output.Summary.Steps
		goals.Goal = output.Goals.Steps

		storeDailyGoals(ctx, goals)
	} else if err != nil {
		log.Errorf(ctx, "error getting item: %v", err)

		return goals, err
	}

	return goals, nil
}

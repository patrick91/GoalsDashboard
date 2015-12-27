package goals

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
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

func GetProfile(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	client, err := getFitbitClient(ctx)

	if err != nil {
		fmt.Fprint(w, "Remember to authenticate to fitbit")

		return
	}

	url := "https://api.fitbit.com/1/user/-/activities/date/2015-12-27.json"

	res, err := client.Get(url)

	if err != nil {
		log.Errorf(ctx, "%v", err)
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	fmt.Fprintf(w, "%s", body)
}

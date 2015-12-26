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

func getFitbitConf(ctx context.Context) (*oauth2.Config, error) {
	var data Settings

	err := GetSettings(ctx, &data)

	if err != nil {
		return nil, err
	}

	var fitbitConf = &oauth2.Config{
		ClientID:     data.FitbitClientID,
		ClientSecret: data.FitbitClientSecret,
		Scopes:       []string{"activity", "weight", "profile"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://www.fitbit.com/oauth2/authorize",
			TokenURL: "https://api.fitbit.com/oauth2/token",
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

func storeToken(ctx context.Context, token *oauth2.Token) error {
	key := datastore.NewKey(ctx, "Tokens", "fitbit", 0, nil)

	_, err := datastore.Put(ctx, key, token)

	if err != nil {
		log.Errorf(ctx, "%v", err)

		return err
	}

	return nil
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
	tok, err := fitbitConf.Exchange(ctx, code)

	if err != nil {
		log.Errorf(ctx, "%v", err)
	}

	storeToken(ctx, tok)

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func GetProfile(w http.ResponseWriter, r *http.Request) {
	var token oauth2.Token
	ctx := appengine.NewContext(r)

	fitbitConf, err := getFitbitConf(ctx)

	if err != nil {
		fmt.Fprint(w, "Remember to initialise your settings")

		return
	}

	err = getToken(ctx, &token)

	if err != nil {
		fmt.Fprint(w, "Remember to authenticate with fitbit")

		return
	}

	// TODO: the token is usually updated automatically,
	// so we need to store the updated one somehow
	client := fitbitConf.Client(ctx, &token)

	url := "https://api.fitbit.com/1/user/-/profile.json"

	res, err := client.Get(url)

	if err != nil {
		log.Errorf(ctx, "%v", err)
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	fmt.Fprintf(w, "%s", body)
}

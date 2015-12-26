package goals

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
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

func init() {
	http.HandleFunc("/", handler)
	http.HandleFunc("/fitbit/auth", redirectToFitbitAuthHandler)
	http.HandleFunc("/fitbit/callback", fitbitAuthCallbackHandler)
	http.HandleFunc("/admin/settings", SettingsHandler)
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "This should be my goals app!")
}

func redirectToFitbitAuthHandler(w http.ResponseWriter, r *http.Request) {
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

func fitbitAuthCallbackHandler(w http.ResponseWriter, r *http.Request) {
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

	// ideally store the token, but now we are going to make a simple test
	// userID := tok.Extra("user_id")

	url := fmt.Sprintf("https://api.fitbit.com/1/user/-/profile.json")

	client := fitbitConf.Client(ctx, tok)

	res, err := client.Get(url)

	if err != nil {
		log.Errorf(ctx, "%v", err)
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	fmt.Fprintf(w, "%s", body)
}

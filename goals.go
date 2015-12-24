package goals

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"google.golang.org/appengine"

	"golang.org/x/oauth2"
)

func getFitbitConf() *oauth2.Config {
	var fitbitConf = &oauth2.Config{
		ClientID:     "FILLME",
		ClientSecret: "FILLME",
		Scopes:       []string{"activity", "weight", "profile"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://www.fitbit.com/oauth2/authorize",
			TokenURL: "https://api.fitbit.com/oauth2/token",
		},
	}

	return fitbitConf
}

func init() {
	http.HandleFunc("/", handler)
	http.HandleFunc("/fitbit/auth", redirectToFitbitAuthHandler)
	http.HandleFunc("/fitbit/callback", fitbitAuthCallbackHandler)
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "This should be my goals app!")
}

func redirectToFitbitAuthHandler(w http.ResponseWriter, r *http.Request) {
	fitbitConf := getFitbitConf()
	// Redirect user to consent page to ask for permission
	// for the scopes specified above.
	url := fitbitConf.AuthCodeURL("state", oauth2.AccessTypeOffline)

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func fitbitAuthCallbackHandler(w http.ResponseWriter, r *http.Request) {
	fitbitConf := getFitbitConf()

	ctx := appengine.NewContext(r)
	code := r.URL.Query()["code"][0]

	// Use the authorization code that is pushed to the redirect URL.
	// NewTransportWithCode will do the handshake to retrieve
	// an access token and initiate a Transport that is
	// authorized and authenticated by the retrieved token.
	tok, err := fitbitConf.Exchange(ctx, code)

	if err != nil {
		log.Fatal(err)
	}

	// ideally store the token, but now we are going to make a simple test
	// userID := tok.Extra("user_id")

	url := fmt.Sprintf("https://api.fitbit.com/1/user/-/profile.json")

	client := fitbitConf.Client(ctx, tok)

	res, err := client.Get(url)

	if err != nil {
		log.Fatal(err)
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	fmt.Fprintf(w, "%s", body)
}

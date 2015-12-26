package goals

import (
	"net/http"
	"text/template"

	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

const settingsForm = `
<html>
	<body>
		<form action="/admin/settings" method="post">
			<div><input value="{{.FitbitClientID}}" name="fitbit_client_id" placeholder="Fitbit Client ID"></div>
			<div><input value="{{.FitbitClientSecret}}" name="fitbit_client_secret" placeholder="Fitbit Client Secret"></div>
			<div><input type="submit" value="Update"></div>
		</form>
	</body>
</html>
`

var settingsFormTemplate = template.Must(template.New("settings").Parse(settingsForm))

// Settings stores the global Application settings
type Settings struct {
	FitbitClientID     string
	FitbitClientSecret string
}

// GetSettings is a wrapper around datastore.Get
func GetSettings(ctx context.Context, data *Settings) error {
	key := datastore.NewKey(ctx, "Settings", "main", 0, nil)

	err := datastore.Get(ctx, key, data)

	return err
}

// SettingsHandler saves the settings
func SettingsHandler(w http.ResponseWriter, r *http.Request) {
	var data Settings

	ctx := appengine.NewContext(r)
	key := datastore.NewKey(ctx, "Settings", "main", 0, nil)

	if r.Method == "POST" {
		fitbitClientID := r.FormValue("fitbit_client_id")
		fitbitClientSecret := r.FormValue("fitbit_client_secret")

		data = Settings{
			FitbitClientID:     fitbitClientID,
			FitbitClientSecret: fitbitClientSecret,
		}

		_, err := datastore.Put(ctx, key, &data)

		if err != nil {
			log.Errorf(ctx, "%v", err)

			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		err := GetSettings(ctx, &data)

		if err != nil {

			if err != datastore.ErrNoSuchEntity {
				log.Errorf(ctx, "%v", err)

				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	}

	settingsFormTemplate.Execute(w, data)
}

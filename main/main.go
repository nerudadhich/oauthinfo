package main

import (
	"context"
	"flag"
	"fmt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

var conf *oauth2.Config

func login() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conf = &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes: []string{
				userInfoScope,
			},
			Endpoint: google.Endpoint,
		}

		req, err := http.NewRequest("GET", conf.AuthCodeURL("testing"), nil)
		if err != nil {
			panic(err)
		}

		client := http.Client{}
		resp, err3 := client.Do(req)
		if err3 != nil {
			panic(err3)
		}

		bodyb, _ := ioutil.ReadAll(resp.Body)

		w.Write(bodyb)

	})
}


func oauth() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/oauth" {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		data, err := getUserDataFromGoogle(r.URL.Query().Get("code"))
		if err != nil {
			log.Println(err.Error())
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}
		fmt.Fprintf(w, "UserInfo: %s\n", data)
	})
}

func getUserDataFromGoogle(code string) ([]byte, error) {
	// Use code to get token and get user info from Google.

	token, err := conf.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("code exchange wrong: %s", err.Error())
	}
	response, err := http.Get(oauthGoogleUrlAPI + token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed getting user info: %s", err.Error())
	}
	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed read response: %s", err.Error())
	}
	return contents, nil
}

// Starts the main server.
func startServer(host, port string) error {
	flag.StringVar(&host, "listen-addr", ":"+port, "server listen address")
	flag.Parse()

	logger := log.New(os.Stdout, "http: ", log.LstdFlags)

	router := http.NewServeMux()

	router.Handle("/login", login())
	router.Handle("/oauth", oauth())

	server := &http.Server{
		Addr:    host,
		Handler: router,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Println(fmt.Sprintf("Could not listen on %s: %v\n", host, err))
		return err
	}

	logger.Println("Server started.")

	return nil
}

func main(){
	startServer("localhost", "5000")
}
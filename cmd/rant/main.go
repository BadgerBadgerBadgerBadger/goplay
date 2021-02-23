package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"github.com/prologic/bitcask"
	log "github.com/sirupsen/logrus"

	"badgerbadgerbadgerbadger.dev/goplay/internal/util"
)

var decoder = schema.NewDecoder()
var indexHtml []byte

var slack Slack
var config Config

func main() {

	rand.Seed(time.Now().UnixNano())

	configPath := flag.String("config-path", "", "provide path to the json config file")
	flag.Parse()

	if *configPath == "" {
		util.Must(errors.New("must provide a config path"))
	}

	if err := util.ConfigFromJsonFile(*configPath, &config); err != nil {
		util.Must(err, "failed to load config")
	}
	log.Infof("config \n%+v\n", config)

	var err error
	db, err = bitcask.Open(config.Database.Path)
	util.Must(err, fmt.Sprintf("failed to load db at path %s", config.Database.Path))

	slack = NewSlack(config.Slack)

	indexHtml, err = ioutil.ReadFile("/home/rant/static/index.html")
	util.Must(err, "failed to read index.html")

	r := mux.NewRouter()

	r.Use(func(next http.Handler) http.Handler {
		return handlers.CombinedLoggingHandler(os.Stdout, next)
	})

	r.HandleFunc("/", indexFileHandler)
	r.HandleFunc("/oauth", oauthHandler)

	// Routes consist of a path and a handler function.
	r.HandleFunc("/rant", rantHandler)

	// Bind to a port and pass our router in
	log.Fatal(http.ListenAndServe(":8000", r))
}

func indexFileHandler(w http.ResponseWriter, r *http.Request) {
	w.Write(indexHtml)
}

func oauthHandler(w http.ResponseWriter, r *http.Request) {

	query := r.URL.Query()

	codeQ, ok := query["code"]
	if !ok || codeQ[0] == "" {
		http.Error(w, "no code available", http.StatusBadRequest)
		return
	}

	err := slack.Authenticate(codeQ[0])

	if err != nil {
		log.Warn(err.Error())
		w.WriteHeader(500)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("https://%s/success.html", config.Host), 301)
}

func rantHandler(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	util.Must(err, "failed to parse form data")

	sc := SlashCommand{}

	err = decoder.Decode(&sc, r.PostForm)
	if err != nil {
		http.Error(w, "Form could not be decoded", http.StatusBadRequest)
		log.WithError(err).Warn("Form could not be decoded")
		return
	}

	// we'll send a response via the response url
	err = slack.Rant(sc)
	if err != nil {
		http.Error(w, "oops, couldn't process that", http.StatusBadRequest)
		log.WithError(err).Warn("oops, couldn't process that")
		return
	}
}

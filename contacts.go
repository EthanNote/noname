// Copyroght (c) 2019, Tencent Inc. All rights reserved.
// auuthor GUO,ZHONGJIE (authurguo@tencent.com)

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"
)

// Contacts is a contacts.
type Contacts struct {
	Name        string `json:"name"`
	Department  string `json:"department"`
	Title       string `json:"title"`
	PhoneNumber string `json:"phoneNumber"`
	Email       string `json:"email"`
}

// Store is a contacts store.
type Store struct {
	Contacts []*Contacts `json:"contacts"`
}

// NewStore return a store.
func NewStore() *Store {
	data := `{
		"contacts": [{
			"name": "郭仲杰",
			"department": "公司其他组织/TME商业广告部",
			"title": "员工",
			"phoneNumber": "0755-86013388-75789",
			"email": "authurguo@tencent.com"
		}]
	}`
	s := &Store{}
	err := json.Unmarshal([]byte(data), s)
	if err != nil {
		panic(err)
	}
	return s
}

// Service is a contacts service.
type Service struct {
	Store *Store
	mutex sync.Mutex
}

// Get is the Service GET method handler.
func (s *Service) Get(w http.ResponseWriter, r *http.Request) {
	if s != nil {
		if s.Store != nil {
			resp, err := json.Marshal(s.Store)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				w.Header().Set("Content-Type", "application/json")
				w.Write(resp)
			}
		}
	}
}

// Post is the Service POST method handler.
func (s *Service) Post(w http.ResponseWriter, r *http.Request) {
	content, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	c := &Contacts{}
	if err = json.Unmarshal(content, c); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if s != nil {
		if s.Store != nil {
			s.mutex.Lock()
			defer s.mutex.Unlock()
			s.Store.Contacts = append(s.Store.Contacts, c)
		}
	}
}

func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s\n", r.Method, r.URL)
	switch r.Method {
	case "GET":
		s.Get(w, r)
	case "POST":
		s.Post(w, r)
	default:
		w.WriteHeader(http.StatusForbidden)
	}
}

// NewService return a new service.
func NewService() *Service {
	return &Service{
		Store: NewStore(),
	}
}

// App contains a HTTP server.
type App struct {
	server *http.Server
}

// Server returns the server in the App.
func (app *App) Server() *http.Server {
	if app != nil {
		return app.server
	}
	return nil
}

// Handle the handler.
func (app *App) Handle(pattern string, handler http.Handler) {
	if app.Server() != nil {
		app.Server().Handler.(*http.ServeMux).Handle(pattern, handler)
	}
}

// HandleFunc the handler func.
func (app *App) HandleFunc(pattern string, handlerFunc http.HandlerFunc) {
	if app.Server() != nil {
		app.Server().Handler.(*http.ServeMux).HandleFunc(pattern, handlerFunc)
	}
}

// Start the app server.
func (app *App) Start() error {
	if app.Server() != nil {
		return app.Server().ListenAndServe()
	}
	return fmt.Errorf("start app with nil server")
}

// Stop the app server.
func (app *App) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if app.Server() != nil {
		return app.Server().Shutdown(ctx)
	}

	return nil
}

// Run the app server until recived a interrupt singal.
func (app *App) Run(address string) {
	if app.Server() != nil {
		app.Server().Addr = address
	}

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)

	go func() {
		<-quit
		if err := app.Stop(); err != nil {
			panic(err)
		}
		defer close(quit)
	}()

	if err := app.Start(); err != nil {
		panic(err)
	}
}

// NewApp return a new app.
func NewApp() *App {
	return &App{
		server: &http.Server{
			Handler: http.NewServeMux(),
		},
	}
}

func main() {
	app := NewApp()
	app.Handle("/contacts", NewService())
	app.Run(":9000")
}

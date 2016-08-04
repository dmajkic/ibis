package ibis

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/dmajkic/ibis/jsonapi"
	// By default, none driver is alway present
	_ "github.com/dmajkic/ibis/jsonapi/none"

	"reflect"

	"github.com/gin-gonic/gin"
)

// Config struct for basic configuration
type Config struct {
	Server         string
	Port           string
	DbURL          string
	DbAdapter      string
	Stderr, Stdout string
}

// Server is core struct
type Server struct {
	*Config
	sync.RWMutex
	Listener net.Listener
	Db       jsonapi.Database
	ModelDb  jsonapi.Database

	exit      chan struct{}
	authToken string
	Tokens    map[string]string
	stopping  bool

	App           interface{}
	AppRouter     AppRouter
	AppAuthorizer AppAuthorizer
}

// StartServer is non-blocking server bootstrap. The actual work is async.
// This is a helper for simple daemon or service support
func (s *Server) StartServer() error {

	// Reload config file
	var config *Config
	var err error

	if config, err = LoadConfig(); err != nil {
		return err
	}

	s.Config = config

	go s.run()
	return nil
}

// Actual server bootstrap proc
func (s *Server) run() {

	var err error

	// Database connection
	err = s.Db.ConnectDB(map[string]string{
		"adapter": s.DbAdapter,
		"dbUrl":   s.DbURL,
	})

	if err != nil {
		log.Printf("%v", err)
		return
	}

	// Web server
	s.Tokens = make(map[string]string)
	s.Listener, err = net.Listen("tcp", s.Server+":"+s.Port)

	if err != nil {
		log.Fatal(err)
		return
	}

	// Router
	router := gin.Default()

	s.SetMiddleware(router)
	s.AppRouter.SetRoutes(router)

	http.Handle("/", router)

	srv := &http.Server{
		Addr: s.Port,
	}

	// Serve somting
	err = srv.Serve(s.Listener)
	if err != nil {
		log.Printf("%v", err)
	}

	return
}

// StopServer closes server listener and stops running server.
func (s *Server) StopServer() error {
	if s.Listener != nil {
		s.Listener.Close()
	}

	return nil
}

// ReloadConfig reloads config.json file via server restart
func (s *Server) ReloadConfig() error {
	s.StopServer()
	return s.StartServer()
}

// ListenAndServe loads config.json, starts server and
// waits until serveer stops
func (s *Server) ListenAndServe() error {

	// Load config file
	var config *Config
	var err error

	if config, err = LoadConfig(); err != nil {
		return err
	}

	s.Config = config

	s.run()
	return nil
}

// LoadConfig loads config.json file from same place where the app is
func LoadConfig() (*Config, error) {
	// Default config
	conf := &Config{
		Server:    "",
		Port:      "2828",
		DbURL:     "username:password@hostname/your_database?charset=utf8&parseTime=True&loc=Local",
		DbAdapter: "mysql",
	}

	// Get config.json in same dir where app is
	path, err := filepath.Abs("config.json")

	// If there is no config, use default settings
	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Println("JSON config file not found.")
		log.Printf("Listen on: %v:%v", conf.Server, conf.Port)
		log.Printf("Database:  (none)")
		log.Println("")
		return conf, nil
	}

	// Open config file
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Decode config
	r := json.NewDecoder(f)
	err = r.Decode(&conf)
	if err != nil {
		return conf, err
	}

	log.Printf("Listen on: %v:%v", conf.Server, conf.Port)
	log.Printf("Database:   %v\n", conf.DbAdapter)
	log.Println("")

	return conf, nil
}

// NewServer constructs new server instance
func NewServer(app interface{}) *Server {

	modeldb := jsonapi.NewDatabase("none")

	server := &Server{
		ModelDb: modeldb,
		App:     app,
	}

	v := reflect.ValueOf(app)
	v.Elem().FieldByName("Server").Set(reflect.ValueOf(server))

	if router, ok := app.(AppRouter); ok {
		server.AppRouter = router
	}

	if auth, ok := app.(AppAuthorizer); ok {
		server.AppAuthorizer = auth
	}

	for _, cent := range cents {
		cent.Init(server)
	}

	return server
}

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
	_ "github.com/dmajkic/ibis/jsonapi/none"

	"github.com/gin-gonic/gin"
)

// Interface that must be implemented by user App
type App interface {
	SetRoutes(router *gin.Engine)
	LoginUser(c *gin.Context, user map[string]interface{}) error
}

// Basic config
type Config struct {
	Server         string
	Port           string
	DbUrl          string
	DbAdapter      string
	Stderr, Stdout string
}

// Core Server struct
type Server struct {
	*Config
	sync.RWMutex
	Listener net.Listener
	Db       jsonapi.Database

	exit      chan struct{}
	authToken string
	Tokens    map[string]string

	stopping bool
	App      App
}

// Start should not block. Do the actual work async.
// This is to allow easy daemon or service support
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
	s.Db, err = s.Db.ConnectDB(map[string]string{
		"adapter": s.DbAdapter,
		"dbUrl":   s.DbUrl,
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
	s.App.SetRoutes(router)

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

// Stop service should be quick
func (s *Server) StopServer() error {
	if s.Listener != nil {
		s.Listener.Close()
	}

	return nil
}

// Reloads config.json file via server restart
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

// Load config.json file from same place where app is
func LoadConfig() (*Config, error) {
	// Default config
	conf := &Config{
		Server:    "",
		Port:      "2828",
		DbUrl:     "username:password@hostname/your_database?charset=utf8&parseTime=True&loc=Local",
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

// Server constructor
func New(app App, ormDriver string) (*Server, error) {

	db, err := jsonapi.NewDatabase(ormDriver)
	if err != nil {
		return nil, err
	}

	return &Server{
		Db:  db,
		App: app,
	}, nil
}

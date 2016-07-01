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
func (p *Server) StartServer() error {

	// Reload config file
	var config *Config
	var err error

	if config, err = LoadConfig(); err != nil {
		return err
	}

	p.Config = config

	go p.run()
	return nil
}

// Actual server bootstrap proc
func (p *Server) run() {

	var err error

	// Database connection
	p.Db, err = p.Db.ConnectDB(map[string]string{
		"adapter": p.DbAdapter,
		"dbUrl":   p.DbUrl,
	})

	if err != nil {
		log.Printf("%v", err)
		return
	}

	// Web server
	p.Tokens = make(map[string]string)
	p.Listener, err = net.Listen("tcp", p.Server+":"+p.Port)
	if err != nil {
		log.Fatal(err)
		//logger.Error(err)
		return
	}

	s := &http.Server{
		Addr: p.Port,
	}

	// Router
	router := gin.Default()

	p.SetMiddleware(router)
	p.App.SetRoutes(router)

	http.Handle("/", router)

	// Serve somting
	err = s.Serve(p.Listener)
	if err != nil {
		log.Printf("%v", err)
	}

	return
}

// Stop service should be quick
func (p *Server) StopServer() error {
	if p.Listener != nil {
		p.Listener.Close()
	}

	return nil
}

// Reloads config.json file via server restart
func (p *Server) ReloadConfig() error {
	p.StopServer()
	return p.StartServer()
}

// ListenAndServe loads config.json, starts server and
// waits until serveer stops
func (p *Server) ListenAndServe() error {

	// Load config file
	var config *Config
	var err error

	if config, err = LoadConfig(); err != nil {
		return err
	}

	p.Config = config

	p.run()
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
		log.Println("JSON config file not found. Using default settings:")
		log.Printf("Listen on: %v:%v", conf.Server, conf.Port)
		log.Printf("Database:   %v\n", conf.DbAdapter)
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

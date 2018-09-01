package main

import (
	"github.com/buaazp/fasthttprouter"
	"github.com/labstack/gommon/log"
	"github.com/namsral/flag"
	"github.com/valyala/fasthttp"
	"time"
)

type NodeStatus struct {
	WSRepStatus   int
	RWEnabled     bool
	NodeAvailable bool
	Timestamp     int
}

type Config struct {
	WebListen        string
	WebReadTimeout   int
	WebWriteTimeout  int
	CheckROEnabled   bool
	CheckInterval    int
	CheckFailTimeout int
	CheckForceEnable bool
	MysqlHost        string
	MysqlPort        int
	MysqlUser        string
	MysqlPass        string
	MysqlTimeout     int
}

var (
	status = &NodeStatus{}
	config *Config
)

func main() {
	var err error
	config, err = parseFlags()
	if err != nil {
		log.Fatalf("Options parsing failed with err: %s", err)
	}

	go checker(status)

	router := getRouter()
	server := &fasthttp.Server{
		Handler:          router.Handler,
		DisableKeepalive: true,
		Concurrency:      2048,
		ReadTimeout:      time.Duration(config.WebReadTimeout) * time.Millisecond,
		WriteTimeout:     time.Duration(config.WebWriteTimeout) * time.Millisecond,
	}

	log.Printf("Server starting on %s", config.WebListen)
	if err := server.ListenAndServe(config.WebListen); err != nil {
		log.Fatalf("Error in ListenAndServe: %s", err)
	}
}

func getRouter() *fasthttprouter.Router {
	router := fasthttprouter.New()
	router.GET("/", checkerHandler)
	router.HEAD("/", checkerHandler)
	return router
}

func parseFlags() (*Config, error) {
	config := Config{}
	flag.StringVar(&config.WebListen, "WEB_LISTEN", ":9200", "Web server listening interface and port")
	flag.IntVar(&config.WebReadTimeout, "WEB_READ_TIMEOUT", 30000, "Web server request read timeout, ms")
	flag.IntVar(&config.WebWriteTimeout, "WEB_WRITE_TIMEOUT", 30000, "Web server request write timeout, ms")
	flag.BoolVar(&config.CheckROEnabled, "CHECK_RO_ENABLED", false, "Mark 'read_only' node as available")
	flag.BoolVar(&config.CheckForceEnable, "CHECK_FORCE_ENABLE", false, "Ignoring the status of the checks and always marking the node as available")
	flag.IntVar(&config.CheckInterval, "CHECK_INTERVAL", 500, "Mysql checks interval, ms")
	flag.IntVar(&config.CheckFailTimeout, "CHECK_FAIL_TIMEOUT", 3000, "Mark the node inaccessible if for the specified time there were no successful checks, ms")
	flag.StringVar(&config.MysqlHost, "MYSQL_HOST", "127.0.0.1", "MySQL host addr")
	flag.IntVar(&config.MysqlPort, "MYSQL_PORT", 3306, "MySQL port")
	flag.StringVar(&config.MysqlUser, "MYSQL_USER", "pxc_checker", "MySQL username")
	flag.StringVar(&config.MysqlPass, "MYSQL_PASS", "", "MySQL password")

	flag.Parse()
	return &config, nil
}

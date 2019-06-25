package main

import (
	log "github.com/sirupsen/logrus"

	"github.com/pkg/errors"
	"github.com/t94j0/satellite/handler"
	"github.com/t94j0/satellite/path"
)

// ProjectName is the current project name
const ProjectName string = "satellite"

func main() {
	log.SetLevel(log.DebugLevel)

	config, err := Config()
	if err != nil {
		log.Fatal(err)
	}

	serverPath := config.GetString("server_path")
	listen := config.GetString("listen")
	certPath := config.GetString("ssl.cert")
	keyPath := config.GetString("ssl.key")
	serverHeader := config.GetString("server_header")
	managementIP := config.GetString("management.ip")
	managementPath := config.GetString("management.path")
	notFoundRedirect := config.GetString("not_found.redirect")
	notFoundRender := config.GetString("not_found.render")
	indexPath := config.GetString("index")
	redirectHTTP := config.GetBool("redirect_http")
	logLevel := config.GetString("log_level")

	logOptions := map[string]log.Level{
		"":      log.DebugLevel,
		"panic": log.PanicLevel,
		"fatal": log.FatalLevel,
		"error": log.ErrorLevel,
		"warn":  log.WarnLevel,
		"info":  log.InfoLevel,
		"debug": log.DebugLevel,
		"trace": log.TraceLevel,
	}

	log.SetLevel(logOptions[logLevel])

	log.Debugf("Using config file %s", config.ConfigFileUsed())
	log.Debugf("Using server path %s", serverPath)

	// Parse .info files
	paths, err := path.New(serverPath)
	if err != nil {
		log.Fatal(err)
	}

	// Warn user about redirection opsec
	if notFoundRedirect == "" && notFoundRender == "" {
		log.Warn("Use not_found handlers for opsec")
	}

	log.Debugf("Loaded %d path(s)", paths.Len())

	// Listen for when files in serverPath change
	go func() {
		if err := createWatcher(serverPath, "1s", func() error {
			return paths.Reload()
		}); err != nil {
			log.Fatal(err)
		}
	}()

	// Create server and listen
	server, err := handler.New(
		paths,
		serverPath,
		listen,
		keyPath,
		certPath,
		serverHeader,
		managementIP,
		managementPath,
		indexPath,
		notFoundRedirect,
		notFoundRender,
		redirectHTTP,
	)
	if err != nil {
		log.Fatal(errors.Wrap(err, "server configuration error"))
	}

	log.Infof("Listening HTTPS on port %s", config.GetString("listen"))
	if err := server.Start(); err != nil {
		log.Fatal(err)
	}
}

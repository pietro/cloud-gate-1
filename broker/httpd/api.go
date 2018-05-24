package httpd

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/Symantec/Dominator/lib/log"
	"github.com/Symantec/cloud-gate/broker"
	"github.com/Symantec/cloud-gate/broker/appconfiguration"
	"github.com/Symantec/cloud-gate/broker/configuration"
)

type HtmlWriter interface {
	WriteHtml(writer io.Writer)
}

type Server struct {
	brokers     map[string]broker.Broker
	appConfig   *appconfiguration.AppConfiguration
	config      *configuration.Configuration
	htmlWriters []HtmlWriter
	logger      log.DebugLogger
}

func StartServer(appConfig *appconfiguration.AppConfiguration, brokers map[string]broker.Broker,
	logger log.DebugLogger) (*Server, error) {
	statusListener, err := net.Listen("tcp", fmt.Sprintf(":%d", appConfig.Base.StatusPort))
	if err != nil {
		return nil, err
	}
	serviceListener, err := net.Listen("tcp", fmt.Sprintf(":%d", appConfig.Base.ServicePort))
	if err != nil {
		return nil, err
	}

	server := &Server{
		brokers:   brokers,
		logger:    logger,
		appConfig: appConfig,
	}
	http.HandleFunc("/", server.rootHandler)
	http.HandleFunc("/status", server.statusHandler)

	serviceMux := http.NewServeMux()
	serviceMux.HandleFunc("/", server.rootHandler)

	statusServer := &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		err := statusServer.Serve(statusListener)
		if err != nil {
			logger.Fatalf("Failed to start status server, err=%s", err)
		}
	}()

	tlsConfig := &tls.Config{
		MinVersion:               tls.VersionTLS12,
		CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
		},
	}
	serviceServer := &http.Server{
		Handler:      serviceMux,
		TLSConfig:    tlsConfig,
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	go func() {
		err := serviceServer.ServeTLS(serviceListener, appConfig.Base.TLSCertFilename, appConfig.Base.TLSKeyFilename)
		if err != nil {
			logger.Fatalf("Failed to start service server, err=%s", err)
		}
	}()

	return server, nil
}

func (s *Server) AddHtmlWriter(htmlWriter HtmlWriter) {
	s.htmlWriters = append(s.htmlWriters, htmlWriter)
}

func (s *Server) UpdateConfiguration(
	config *configuration.Configuration) error {
	s.config = config
	return nil
}

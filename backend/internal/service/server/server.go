package server

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	"github.com/hasmcp/hasmcp-ce/backend/internal/service/config"
	zlog "github.com/rs/zerolog/log"
	"golang.org/x/crypto/acme/autocert"
)

type (
	Service interface {
		fiber.Router

		GetRoutes() []fiber.Route
		Start(ctx context.Context) error
		Shutdown(ctx context.Context) error
	}

	service struct {
		cfg     serverCfg
		acmesrv *http.Server
		*fiber.App
	}

	Params struct {
		Config config.Service
	}

	serverCfg struct {
		Port                string `yaml:"port"`
		SSLPort             string `yaml:"sslPort"`
		SslEnabled          bool   `yaml:"sslEnabled"`
		SslCacheDir         string `yaml:"sslCacheDir"`
		LetsencryptEmail    string `yaml:"letsencryptEmail"`
		DomainName          string `yaml:"domainName"`
		CaseSensitiveRoutes bool   `yaml:"caseSensitiveRoutes"`
		MaxBodySizeInBytes  int    `yaml:"maxBodySizeInBytes"`
	}
)

const (
	_cfgKey = "server"

	_logPrefix = "[httpserver] "
)

var (
	_entityTooLarge = []byte(`{"error": {"message":"max body size", "code":413}}`)
	_domainName     = ""
)

func New(p Params) (Service, error) {
	var cfg serverCfg
	err := p.Config.Populate(_cfgKey, &cfg)
	if err != nil {
		return nil, err
	}

	_domainName = cfg.DomainName

	return &service{
		App: fiber.New(fiber.Config{
			AppName:       p.Config.App() + " " + p.Config.Version() + " (" + p.Config.Env() + ")",
			ServerHeader:  p.Config.AppShortName() + " " + p.Config.Version(),
			CaseSensitive: true,
			BodyLimit:     cfg.MaxBodySizeInBytes,
			ErrorHandler: func(c *fiber.Ctx, err error) error {
				var e *fiber.Error
				if errors.As(err, &e) {
					if e.Code == fiber.StatusRequestEntityTooLarge {
						return c.Status(e.Code).Send(_entityTooLarge)
					}
				}

				return fiber.DefaultErrorHandler(c, err)
			},
		}),
		cfg: cfg,
	}, nil
}

func (s *service) Start(ctx context.Context) error {
	s.App.Use(healthcheck.New()).Use(compress.New())
	zlog.Info().Any("routes", s.GetRoutes()).Msg(_logPrefix + "registered routes")

	if !s.cfg.SslEnabled {
		err := s.App.Listen(":" + s.cfg.Port)
		if err != nil {
			return err
		}
		return nil
	}

	zlog.Info().Msg(_logPrefix + "ssl enabled")
	m := &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(_domainName),
		Cache:      autocert.DirCache(s.cfg.SslCacheDir),
		Email:      s.cfg.LetsencryptEmail,
	}

	go func() {
		// Create a standard HTTP handler that serves the ACME challenge requests
		// and redirects everything else to HTTPS.
		zlog.Info().Str("domainName", _domainName).Str("port", s.cfg.Port).
			Msg(_logPrefix + "starting ACME challenge HTTP listener")

		s.acmesrv = &http.Server{
			Addr:    ":" + s.cfg.Port,
			Handler: m.HTTPHandler(http.HandlerFunc(redirectHTTP)), // nil means default redirect to HTTPS
		}
		if err := s.acmesrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zlog.Fatal().Err(err).Msg(_logPrefix + "ACME HTTP listener failed")
		}
		zlog.Info().Msg(_logPrefix + "ACME HTTP listener shut down")
	}()

	// TLS Config
	tlsConfig := &tls.Config{
		ClientSessionCache: tls.NewLRUClientSessionCache(100),
		// Get Certificate from Let's Encrypt
		GetCertificate: func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
			if hello.ServerName != s.cfg.DomainName {
				return nil, errors.New("unexpected domain")
			}
			// 1. Set up the timeout context
			timeout := 30 * time.Second
			timeoutCtx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			// 2. Channel to receive the result from autocert
			certCh := make(chan *tls.Certificate, 1)
			errCh := make(chan error, 1)

			// 3. Run the blocking GetCertificate call in a goroutine
			go func() {
				cert, err := m.GetCertificate(hello)
				if err != nil {
					errCh <- err
					return
				}
				certCh <- cert
			}()

			// 4. Wait on the context or the result channels
			select {
			case cert := <-certCh:
				return cert, nil

			case err := <-errCh:
				if err.Error() == "acme/autocert: missing server name" {
					return nil, err
				}
				// This is the error from autocert (e.g., Let's Encrypt failing)
				zlog.Error().Err(err).Msg(_logPrefix + "autocert GetCertificate failed (Likely DNS/Firewall issue)")
				return nil, err

			case <-timeoutCtx.Done():
				// This is the manual timeout firing
				err := fmt.Errorf("autocert timed out after %s", timeout)
				zlog.Error().Err(err).Msg(_logPrefix + "autocert timed out while waiting for certificate.")
				return nil, err
			}
		},
		// Secure configuration recommended by Mozilla:
		MinVersion: tls.VersionTLS12,
		CurvePreferences: []tls.CurveID{
			tls.CurveP521,
			tls.CurveP384,
			tls.CurveP256,
		},
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_RSA_WITH_AES_256_CBC_SHA,
		},
		// By default NextProtos contains the "h2"
		// This has to be removed since Fasthttp does not support HTTP/2
		// Or it will cause a flood of PRI method logs
		// http://webconcepts.info/concepts/http-method/PRI
		NextProtos: []string{
			"http/1.1", "acme-tls/1",
		},
	}

	ln, err := tls.Listen("tcp", ":"+s.cfg.SSLPort, tlsConfig)
	if err != nil {
		zlog.Error().Err(err).Msg(_logPrefix + "fiber HTTPS listener failed on tls listen")
		return err
	}

	if err := s.App.Listener(ln); err != nil {
		zlog.Error().Err(err).Msg(_logPrefix + "fiber HTTPS listener failed")
		return err
	}

	return nil
}

func (s *service) Shutdown(ctx context.Context) error {
	if s.acmesrv != nil {
		zlog.Info().Msg(_logPrefix + "fiber listener going down for acme server")
		_ = s.acmesrv.Shutdown(ctx)
	}
	zlog.Info().Msg(_logPrefix + "fiber listener going down")
	defer zlog.Info().Msg(_logPrefix + "fiber listener shutdown")
	return s.App.ShutdownWithContext(ctx)
}

func (s *service) GetRoutes() []fiber.Route {
	return s.App.GetRoutes()
}

func redirectHTTP(w http.ResponseWriter, r *http.Request) {
	// Skip redirect for ACME challenge requests
	if r.URL.Path == "/.well-known/acme-challenge/" {
		return
	}
	if r.Host != _domainName {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(""))
		return
	}
	http.Redirect(w, r, "https://"+r.Host+r.RequestURI, http.StatusMovedPermanently)
}

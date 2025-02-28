package main

import (
	"context"
	"embed"
	"fmt"
	"io/ioutil"
	a "log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/FM1337/flagbrew-website-source-code/pkg/api"
	"github.com/FM1337/flagbrew-website-source-code/pkg/daemon"
	"github.com/FM1337/flagbrew-website-source-code/pkg/helpers"
	"github.com/FM1337/flagbrew-website-source-code/pkg/models"
	mongo "github.com/FM1337/flagbrew-website-source-code/pkg/mongodb"
	"github.com/alexedwards/scs"
	"github.com/apex/log"
	"github.com/getsentry/sentry-go"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"github.com/google/uuid"
	_ "github.com/joho/godotenv/autoload"
	"github.com/lrstanley/chix"
	"github.com/lrstanley/clix"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/oauth2"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// TODO: let the http server also have a flag to initiate things like tmpdir
// cleanup.

type Flags struct {
	Configured bool   `long:"configured" env:"CONFIGURED" required:"true" description:"If set to false, the web application will exit, should be set to true when everything is configured correctly"`
	Env        string `short:"e" long:"env" env:"ENV" required:"true" description:"The environment the program is running in: production/development"`
	Quiet      bool   `short:"q" long:"quiet" description:"disable logger stdout"`
	HTTP       string `short:"b" long:"http" default:":8080" env:"BIND" description:"ip:port pair to bind to" required:"true"`
	SiteURL    string `short:"u" long:"site-url" required:"true" env:"SITE_URL"`
	HomeURL    string `long:"home-url" required:"true" env:"HOME_DOMAIN"`
	CoreAPIURL string `long:"coreapi-url" required:"true" env:"COREAPI_URL"`
	Proxy      bool   `short:"p" long:"behind-proxy" description:"if X-Forwarded-For headers should be trusted"`
	SessionKey string `long:"session-key" env:"SESSION_KEY" description:"HTTP salted session key (change this to logout all users)"`
	TLS        struct {
		Enable bool   `long:"enable" description:"run tls server rather than standard http"`
		Cert   string `long:"cert" description:"path to ssl cert file"`
		Key    string `long:"key" description:"path to ssl key file"`
	} `group:"TLS Options" env-namespace:"TLS"`
	DB struct {
		URI      string `long:"uri" description:"mongodb uri, e.g. hostname[:port][/database]" required:"true" env:"URI"`
		MaxConns int    `long:"max-connections" default:"128" description:"set maximum amount of connections to keep in the pool" required:"true"`
		Username string `long:"db-username" env:"USER" required:"true" description:"The username to use for logging into the database"`
		Password string `long:"db-password" env:"PASS" required:"true" description:"The password to use for logging into the database"`
	} `group:"Database (MongoDB) Options" env-namespace:"DB"`
	Auth struct {
		Github struct {
			ClientID     string `long:"client-id" env:"GITHUB_CLIENT_ID" description:"GitHub OAuth Client ID"`
			ClientSecret string `long:"client-secret" env:"GITHUB_CLIENT_SECRET" description:"GitHub OAuth Client Secret"`
			Admins       []int  `long:"admins" env:"GITHUB_ADMINS" env-delim:"," description:"user id's of the users you want to be admins"`
		} `group:"GitHub Options" namespace:"github"`
	} `group:"Authentication Options" env-namespace:"AUTH"`
	API struct {
		GitHub struct {
			AccessToken string `long:"github-access-token" env:"GITHUB_ACCESS_TOKEN" description:"token used for fetching GitHub info" required:"true"`
		} `group:"API Keys/Secrets" namespace:"API"`
		Sentry struct {
			DSN string `long:"sentry-io-dsn" env:"SENTRY_DSN" description:"The DSN for sentry.io" required:"true"`
		} `group:"API Keys/Secrets" namespace:"API"`
		Neutrino struct {
			Key string `long:"neutrinoapi-key" env:"NEUTRINOAPI_KEY" description:"The API key used for neutrinoapi" required:"true"`
		}
	}
	WebHooks struct {
		Discord string `long:"gpss-discord-webhook" env:"GPSS_DISCORD_WEBHOOK" description:"The webhook to post to for GPSS Uploads for Discord" required:"true"`
	}
	Secrets struct {
		Github struct {
			UploadSecret string `long:"github-upload-secret" env:"GITHUB_UPLOAD_SECRET" description:"secret key used for allowing GitHub to push builds" required:"true"`
		}
	} `group:"Secrets for Webhooks" env-namespace:"SECRETS"`
}

var (
	// For use with goreleaser, it will auto-inject version/commit/date/etc.
	cli = &clix.CLI[Flags]{}

	logger log.Interface
	debug  = a.New(ioutil.Discard, "debug:", a.Lshortfile|a.LstdFlags)

	oauthConfig *oauth2.Config
	session     *scs.Manager
	// Services
	svcUsers    models.UserService
	svcGitHub   models.GitHubService
	svcGPSS     models.GPSSService
	svcLogs     models.LogService
	svcBans     models.BanService
	svcSettings models.SettingService
	svcFile     models.FileService
	svcPatron   models.PatronService
	svcApproval models.ApprovalService
	svcRestrict models.RestrictService
	svcFilter   models.FilterService
	// APIs
	apiGitHub models.GitHubAPI
	apiGPSS   models.GPSSAPI
	// Daemons
	daemonGitHub         models.GitHubDaemon
	daemonGPSSCleanup    models.GPSSCleanupDaemon
	daemonPatreonCleanup models.PatreonCleanupDaemon
	daemonSetting        models.SettingDaemon
	// Settings
	loadedSettings map[string]*models.Setting // If empty upon loading from database, we'll load the defaults
	legacyPort     string                     // remove this when legacy support is done
	//go:embed all:public/dist
	staticFS  embed.FS
	legacyKey string
)

func main() {
	cli.Parse()
	logger = cli.Logger

	if !cli.Flags.Configured {
		fmt.Println("Not configured yet, please configure")
		os.Exit(1)
	}

	// generate a random string
	legacyKey = uuid.New().String()

	session = scs.NewCookieManager(cli.Flags.SessionKey)

	legacyPort = strings.Split(cli.Flags.HTTP, ":")[1]
	helpers.InitSentry(cli.Flags.API.Sentry.DSN, cli.Flags.Env, cli.Debug)
	// init CoreAPI
	if !helpers.InitCoreAPI(cli.Flags.CoreAPIURL) {
		sentry.Flush(2 * time.Second)
		logger.Fatalf("Could not connect to CoreAPI at URL: %s", cli.Flags.CoreAPIURL)
	}
	mongoSrv := mongo.NewSrv()
	dbCloser := mongoSrv.Close
	if err := mongoSrv.Setup(cli.Flags.DB.URI, cli.Flags.DB.Username, cli.Flags.DB.Password, cli.Flags.DB.MaxConns, nil); err != nil {
		helpers.LogToSentry(err)
		sentry.Flush(2 * time.Second)
		logger.Fatalf("error initializing database: %v", err)
	}
	defer dbCloser()

	svcUsers = mongoSrv.NewUserService()
	svcGitHub = mongoSrv.NewGitHubService()
	apiGitHub = api.NewGitHubAPI(cli.Flags.API.GitHub.AccessToken)
	svcLogs = mongoSrv.NewLogService()
	apiGPSS = api.NewGPSSAPI()
	svcGPSS = mongoSrv.NewGPSSService(&svcLogs)
	svcBans = mongoSrv.NewBanService()
	svcSettings = mongoSrv.NewSettingService()
	svcFile = mongoSrv.NewFileService()
	svcPatron = mongoSrv.NewPatronService()
	svcApproval, svcRestrict = mongoSrv.NewApprovalSvc(&svcLogs, &svcGPSS)
	svcFilter = mongoSrv.NewFilterService()

	// Load the settings
	settingsSlice, _, count, _ := svcSettings.ListSettings(context.Background(), bson.M{}, 1, 10000, bson.M{})
	if count == 0 {
		// Load defaults
		if err := svcSettings.LoadDefaults(context.Background()); err != nil {
			helpers.LogToSentry(err)
			sentry.Flush(2 * time.Second)
			logger.Fatalf("error loading default settings: %v", err)
		}
		loadedSettings = helpers.LoadSettings(nil)
	} else {
		loadedSettings = helpers.LoadSettings(settingsSlice)
	}

	// Init daemons
	daemonGitHub = daemon.NewGitHubDaemon(&apiGitHub, &svcGitHub)
	daemonGPSSCleanup = daemon.NewGPSSCleanupDaemon(loadedSettings, &svcGPSS, &svcLogs, &svcSettings)
	daemonPatreonCleanup = daemon.NewPatreonCleanupDaemon(&svcFile, &svcLogs)
	daemonSetting = daemon.NewSettingDaemon(cli.Flags.HomeURL)

	// Start the daemons
	daemonGPSSCleanup.Start()
	daemonPatreonCleanup.Start()
	daemonSetting.Start()

	oauthConfig = &oauth2.Config{
		ClientID:     cli.Flags.Auth.Github.ClientID,
		ClientSecret: cli.Flags.Auth.Github.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://github.com/login/oauth/authorize",
			TokenURL: "https://github.com/login/oauth/access_token",
		},
		RedirectURL: "",
		Scopes:      []string{}, // TODO?
	}

	r := chi.NewRouter()

	r.Use(
		chix.UseContextIP,
		middleware.RequestID,
		chix.UseStructuredLogger(logger),
		chix.UseDebug(cli.Debug),
		chix.Recoverer,
		middleware.StripSlashes,
		middleware.Compress(5),
		middleware.Maybe(middleware.StripSlashes, func(r *http.Request) bool {
			return !strings.HasPrefix(r.URL.Path, "/debug/")
		}),
		chix.UseNextURL,
		middleware.Timeout(30*time.Second),
	)

	r.Use(
		chix.UseHeaders(map[string]string{
			"Content-Security-Policy":          "default-src 'self'; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; img-src *; font-src 'self' https://fonts.gstatic.com; media-src *; object-src 'none'; child-src 'none'; frame-src 'none'; worker-src 'none'",
			"X-Frame-Options":                  "DENY",
			"X-Content-Type-Options":           "nosniff",
			"Referrer-Policy":                  "no-referrer-when-downgrade",
			"Permissions-Policy":               "clipboard-write=(self)",
			"Access-Control-Allow-Origin":      ".flagbrew.org",
			"Access-Control-Allow-Methods":     "GET, POST, OPTIONS, PUT, DELETE, PATCH",
			"Access-Control-Allow-Headers":     "Origin, X-Requested-With, Content-Type, Accept, Authorization",
			"Access-Control-Allow-Credentials": "true",
		}),
		// auth.AddToContext,
		httprate.LimitByIP(200, 1*time.Minute),
	)

	r.Use(
		chix.UseSecurityTxt(&chix.SecurityConfig{
			ExpiresIn: 182 * 24 * time.Hour,
			Contacts: []string{
				"[REDACTED]",
				"https://github.com/FM1337",
			},
			KeyLinks:  []string{"[REDACTED]"},
			Languages: []string{"en"},
		}),
	)

	if !cli.Debug {
		daemonGitHub.Start()
	}
	r.With(banCheck, emegerencyMode).NotFound(chix.UseStatic(context.Background(), &chix.Static{
		FS:         staticFS,
		CatchAll:   true,
		AllowLocal: cli.Debug,
		Path:       "public/dist",
		Prefix:     "/static/dist",
		SPA:        true,
		Headers: map[string]string{
			"Vary":          "Accept-Encoding",
			"Cache-Control": "public, max-age=7776000",
		},
	}).ServeHTTP)

	// register the metrics
	helpers.RegisterMetrics(svcGPSS)
	registerHTTPRoutes(r)

	// Setup our http server.
	srv := &http.Server{
		Addr:    cli.Flags.HTTP,
		Handler: session.Use(r),

		// Some sane defaults.
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	done := make(chan struct{})

	go func() {
		if cli.Flags.TLS.Enable {
			logger.Infof("initializing https server on %s", cli.Flags.HTTP)
			if err := srv.ListenAndServeTLS(cli.Flags.TLS.Cert, cli.Flags.TLS.Key); err != nil {
				if err != http.ErrServerClosed {
					logger.Fatalf("error in http server: %v", err)
				}
			}
		} else {
			logger.Infof("initializing http server on %s", cli.Flags.HTTP)
			if err := srv.ListenAndServe(); err != nil {
				if err != http.ErrServerClosed {
					logger.Fatalf("error in http server: %v", err)
				}
			}
		}
		close(done)
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	select {
	case <-signals:
		logger.Info("received SIGINT/SIGTERM/SIGQUIT, closing connections...")
		// Stop all daemons
		daemonGitHub.Stop()
		daemonGPSSCleanup.Stop()
		// Run close functions.
		srv.Close()
		dbCloser()
		defer sentry.Flush(2 * time.Second)

		logger.Info("done cleaning up; exiting")
		os.Exit(1)
	case <-done:
	}
}

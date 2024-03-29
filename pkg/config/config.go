package config

import (
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

type Config struct {
	Development        bool
	UseHttps           bool
	GenCerts           bool
	DNSNames           []string
	CertFile           string
	KeyFile            string
	Port               int
	ClientAddresses    []*url.URL
	DataDir            string
	SpotifyID          string
	SpotifySecret      string
	SpotifyRedirectURL string
	RedisAddress       string
	RedisDatabase      string
	RedisPassword      string
	CookieSameSite     http.SameSite
	CookieSecure       bool
}

func New() *Config {
	// Set some default configs
	clientAddress, _ := url.Parse("http://localhost:4200")
	clientAddresses := []*url.URL{clientAddress}
	c := &Config{
		Development:     false,
		UseHttps:        false,
		GenCerts:        false,
		CertFile:        "./data/cert.pem",
		KeyFile:         "./data/key.pem",
		Port:            3000,
		ClientAddresses: clientAddresses,
		DataDir:         "./data",
		RedisAddress:    "localhost:6379",
		RedisDatabase:   "0",
		RedisPassword:   "",
		CookieSameSite:  http.SameSiteLaxMode,
		CookieSecure:    true,
	}

	// Set c.LogLevel
	var logLevel log.Level
	logLevelVal, _ := log.ParseLevel(os.Getenv("JAM_LOG_LEVEL"))
	if logLevelVal != logLevel {
		log.SetLevel(logLevelVal)
	} else {
		log.Debug("Failed to parse JAM_LOG_LEVEL. Using ", log.GetLevel())
	}

	// Set Development mode
	useDevelopmentModeVal := os.Getenv("JAM_DEVELOPMENT")
	if useDevelopmentModeVal != "" {
		useDevelopment, err := strconv.ParseBool(useDevelopmentModeVal)
		if err != nil {
			log.Fatal("Failed to parse JAM_DEVELOPMENT: ", err)
		}
		c.Development = useDevelopment
	} else {
		log.Debug("JAM_DEVELOPMENT is empty. Using ", c.Development)
	}
	if c.Development {
		log.Warn("JAM FACTORY IS RUNNING IN DEVELOPMENT MODE!")
	}

	// Set c.DataDir
	dataDirVal := os.Getenv("JAM_DATA_DIR")
	if dataDirVal != "" {
		c.DataDir = dataDirVal
		c.CertFile = path.Join(c.DataDir, "cert.pem")
		c.KeyFile = path.Join(c.DataDir, "key.pem")
	} else {
		log.Debug("JAM_DATA_DIR is empty. Using ", c.DataDir)
	}

	// Set Cookie related settings
	useCookieSecureVal := os.Getenv("JAM_COOKIE_SECURE")
	if useCookieSecureVal != "" {
		useCookieSecure, err := strconv.ParseBool(useCookieSecureVal)
		if err != nil {
			log.Fatal("Failed to parse JAM_COOKIE_SECURE: ", err)
		}
		c.CookieSecure = useCookieSecure
	} else {
		log.Debug("JAM_COOKIE_SECURE is empty. Using ", c.CookieSecure)
	}

	// Set HTTPS related settings
	useHttpsVal := os.Getenv("JAM_USE_HTTPS")
	if useHttpsVal != "" {
		useHttps, err := strconv.ParseBool(useHttpsVal)
		if err != nil {
			log.Fatal("Failed to parse JAM_USE_HTTPS: ", err)
		}
		c.UseHttps = useHttps
	} else {
		log.Debug("JAM_USE_HTTPS is empty. Using ", c.UseHttps)
	}
	if c.UseHttps {
		// Set c.GenCerts value
		genCertsVal := os.Getenv("JAM_GEN_CERTS")
		if genCertsVal != "" {
			genCerts, err := strconv.ParseBool(genCertsVal)
			if err != nil {
				log.Fatal("Failed to parse JAM_GEN_CERTS: ", err)
			} else {
				c.GenCerts = genCerts
			}
		}

		if c.GenCerts {
			dnsNamesVal := os.Getenv("JAM_DNS_NAMES")
			if dnsNamesVal == "" {
				log.Fatal("JAM_DNS_NAMES cannot be empty when JAM_GEN_CERTS is true")
			}
			dnsNames := strings.Split(dnsNamesVal, ",")
			c.DNSNames = dnsNames
		} else {
			// Set c.CertFile value
			certFileVal := os.Getenv("JAM_CERT_FILE")
			if certFileVal != "" {
				c.CertFile = certFileVal
			} else {
				log.Debug("JAM_CERT_FILE is empty. Using ", c.CertFile)
			}
			// Set c.KeyFile value
			keyFileVal := os.Getenv("JAM_KEY_FILE")
			if keyFileVal != "" {
				c.KeyFile = keyFileVal
			} else {
				log.Debug("JAM_KEY_FILE is empty. Using ", c.KeyFile)
			}
		}
	}

	// Set c.Port
	portVal := os.Getenv("JAM_PORT")
	if portVal != "" {
		port, err := strconv.Atoi(portVal)
		if err != nil {
			log.Fatal("failed to parse JAM_PORT: ", err)
		}
		c.Port = port
	} else {
		log.Debug("JAM_PORT is empty. Using ", c.Port)
	}

	// Set c.ClientAddresses
	clientAddressesVal := os.Getenv("JAM_CLIENT_ADDRESSES")
	if clientAddressesVal != "" {
		clientAddressArr := strings.Split(strings.Replace(clientAddressesVal, " ", "", -1), ",")
		clientAddresses := make([]*url.URL, len(clientAddressArr))
		for i := range clientAddressArr {
			url, err := url.Parse(clientAddressArr[i])
			if err != nil {
				log.Fatal("failed to parse JAM_CLIENT_ADDRESSES: ", err)
			}
			clientAddresses[i] = url
		}
		c.ClientAddresses = clientAddresses
	} else {
		log.Debug("JAM_CLIENT_ADDRESSES is empty. Using ", c.ClientAddresses)
	}

	// Set c.RedisAddress
	redisAddressVal := os.Getenv("JAM_REDIS_ADDRESS")
	if redisAddressVal != "" {
		c.RedisAddress = redisAddressVal
	} else {
		log.Debug("JAM_REDIS_ADDRESS is empty. Using: ", c.RedisAddress)
	}

	// Set c.RedisDatabase
	redisDatabaseVal := os.Getenv("JAM_REDIS_DATABASE")
	if redisDatabaseVal != "" {
		c.RedisDatabase = redisDatabaseVal
	} else {
		log.Debug("JAM_REDIS_DATABASE is empty. Using ", c.RedisDatabase)
	}

	// Set c.RedisPassword
	c.RedisPassword = os.Getenv("JAM_REDIS_PASSWORD")

	// Set c.Spotify* values
	c.SpotifyID = os.Getenv("JAM_SPOTIFY_ID")
	if c.SpotifyID == "" {
		log.Fatal("JAM_SPOTIFY_ID cannot be empty")
	}
	c.SpotifySecret = os.Getenv("JAM_SPOTIFY_SECRET")
	if c.SpotifySecret == "" {
		log.Fatal("JAM_SPOTIFY_SECRET cannot be empty")
	}
	c.SpotifyRedirectURL = os.Getenv("JAM_SPOTIFY_REDIRECT_URL")
	if c.SpotifyRedirectURL == "" {
		log.Fatal("JAM_SPOTIFY_REDIRECT_URL cannot be empty")
	}

	return c
}

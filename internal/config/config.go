package config

import (
	"crypto/sha256"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/USA-RedDragon/aredn-manager/internal/utils"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/pbkdf2"
)

// Config stores the application configuration.
type Config struct {
	Debug                             bool
	Port                              int
	PasswordSalt                      string
	OTLPEndpoint                      string
	InitialAdminUserPassword          string
	CORSHosts                         []string
	TrustedProxies                    []string
	HIBPAPIKey                        string
	ServerName                        string
	Supernode                         bool
	Masquerade                        bool
	WireguardTapAddress               string
	NodeIP                            string
	strSessionSecret                  string
	SessionSecret                     []byte
	VTUNStartingAddress               string
	WireguardStartingAddress          string
	WireguardStartingPort             uint16
	PostgresDSN                       string
	postgresUser                      string
	postgresPassword                  string
	postgresHost                      string
	postgresPort                      int
	postgresDatabase                  string
	MetricsNodeExporterHost           string
	MetricsPort                       int
	Latitude                          string
	Longitude                         string
	Gridsquare                        string
	DisableVTun                       bool
	AdditionalOlsrdInterfaces         []string
	AdditionalOlsrdInterfacesIsolated []string
}

func loadConfig() Config {
	portStr := os.Getenv("HTTP_PORT")
	httpPort, err := strconv.ParseInt(portStr, 10, 0)
	if err != nil {
		httpPort = int64(3333)
	}

	portStr = os.Getenv("PG_PORT")
	pgPort, err := strconv.ParseInt(portStr, 10, 0)
	if err != nil {
		pgPort = 0
	}

	portStr = os.Getenv("METRICS_PORT")
	metricsPort, err := strconv.ParseInt(portStr, 10, 0)
	if err != nil {
		metricsPort = 0
	}

	portStr = os.Getenv("WIREGUARD_STARTING_PORT")
	wireguardStartingPort, err := strconv.ParseInt(portStr, 10, 0)
	if err != nil {
		wireguardStartingPort = 5527
	}

	tmpConfig := Config{
		Debug:                             os.Getenv("DEBUG") != "",
		Port:                              int(httpPort),
		PasswordSalt:                      os.Getenv("PASSWORD_SALT"),
		OTLPEndpoint:                      os.Getenv("OTLP_ENDPOINT"),
		InitialAdminUserPassword:          os.Getenv("INIT_ADMIN_USER_PASSWORD"),
		HIBPAPIKey:                        os.Getenv("HIBP_API_KEY"),
		ServerName:                        strings.ToUpper(os.Getenv("SERVER_NAME")),
		Supernode:                         os.Getenv("SUPERNODE") != "",
		Masquerade:                        os.Getenv("MASQUERADE") != "",
		WireguardTapAddress:               os.Getenv("WIREGUARD_TAP_ADDRESS"),
		NodeIP:                            os.Getenv("NODE_IP"),
		strSessionSecret:                  os.Getenv("SESSION_SECRET"),
		VTUNStartingAddress:               os.Getenv("VTUN_STARTING_ADDRESS"),
		WireguardStartingAddress:          os.Getenv("WIREGUARD_STARTING_ADDRESS"),
		WireguardStartingPort:             uint16(wireguardStartingPort),
		postgresUser:                      os.Getenv("PG_USER"),
		postgresPassword:                  os.Getenv("PG_PASSWORD"),
		postgresHost:                      os.Getenv("PG_HOST"),
		postgresPort:                      int(pgPort),
		postgresDatabase:                  os.Getenv("PG_DATABASE"),
		MetricsNodeExporterHost:           os.Getenv("METRICS_NODE_EXPORTER_HOST"),
		MetricsPort:                       int(metricsPort),
		Latitude:                          os.Getenv("SERVER_LAT"),
		Longitude:                         os.Getenv("SERVER_LON"),
		Gridsquare:                        os.Getenv("SERVER_GRIDSQUARE"),
		DisableVTun:                       os.Getenv("DISABLE_VTUN") != "",
		AdditionalOlsrdInterfaces:         strings.Split(os.Getenv("ADDITIONAL_OLSRD_INTERFACES"), ","),
		AdditionalOlsrdInterfacesIsolated: strings.Split(os.Getenv("ADDITIONAL_OLSRD_INTERFACES_ISOLATED"), ","),
	}

	if tmpConfig.VTUNStartingAddress == "" {
		panic("VTUN_STARTING_ADDRESS not set")
	}

	if tmpConfig.WireguardStartingAddress == "" {
		panic("WIREGUARD_STARTING_ADDRESS not set")
	}

	if net.ParseIP(tmpConfig.WireguardStartingAddress) == nil {
		panic("WIREGUARD_STARTING_ADDRESS is not a valid IP address")
	}

	if net.ParseIP(tmpConfig.VTUNStartingAddress) == nil {
		panic("VTUN starting address is not a valid IP address")
	}

	if tmpConfig.InitialAdminUserPassword == "" {
		fmt.Println("Initial admin user password not set, using auto-generated password")
		const randLen = 15
		const randNums = 4
		const randSpecial = 2
		tmpConfig.InitialAdminUserPassword, err = utils.RandomPassword(randLen, randNums, randSpecial)
		if err != nil {
			fmt.Println("Password generation failed")
			os.Exit(1)
		}
	}

	if tmpConfig.ServerName == "" {
		panic("Server name not set")
	}

	if tmpConfig.NodeIP == "" {
		panic("Node IP not set")
	}

	if tmpConfig.PasswordSalt == "" {
		tmpConfig.PasswordSalt = "salt"
		fmt.Println("Password salt not set, using INSECURE default")
	}

	if tmpConfig.postgresUser == "" {
		tmpConfig.postgresUser = "postgres"
	}

	if tmpConfig.postgresPassword == "" {
		tmpConfig.postgresPassword = "password"
	}

	if tmpConfig.postgresHost == "" {
		tmpConfig.postgresHost = "localhost"
	}

	if tmpConfig.postgresPort == 0 {
		tmpConfig.postgresPort = 5432
	}

	if tmpConfig.postgresDatabase == "" {
		tmpConfig.postgresDatabase = "postgres"
	}

	if tmpConfig.MetricsNodeExporterHost == "" {
		tmpConfig.MetricsNodeExporterHost = "node-exporter"
	}

	tmpConfig.PostgresDSN = "host=" + tmpConfig.postgresHost + " port=" + strconv.FormatInt(int64(tmpConfig.postgresPort), 10) + " user=" + tmpConfig.postgresUser + " dbname=" + tmpConfig.postgresDatabase + " password=" + tmpConfig.postgresPassword

	// CORS_HOSTS is a comma separated list of hosts that are allowed to access the API
	corsHosts := os.Getenv("CORS_HOSTS")
	if corsHosts == "" {
		tmpConfig.CORSHosts = []string{
			fmt.Sprintf("http://localhost:%d", tmpConfig.Port),
			fmt.Sprintf("http://127.0.0.1:%d", tmpConfig.Port),
		}
	} else {
		tmpConfig.CORSHosts = strings.Split(corsHosts, ",")
	}
	trustedProxies := os.Getenv("TRUSTED_PROXIES")
	if trustedProxies == "" {
		tmpConfig.TrustedProxies = []string{}
	} else {
		tmpConfig.TrustedProxies = strings.Split(trustedProxies, ",")
	}

	const iterations = 4096
	const keyLen = 32
	tmpConfig.SessionSecret = pbkdf2.Key([]byte(tmpConfig.strSessionSecret), []byte(tmpConfig.PasswordSalt), iterations, keyLen, sha256.New)

	return tmpConfig
}

// GetConfig obtains the current configuration
func GetConfig(cmd *cobra.Command) *Config {
	currentConfig := loadConfig()

	// Override with command line flags
	if cmd != nil {
		debug, err := cmd.Flags().GetBool("debug")
		if err == nil && debug {
			currentConfig.Debug = debug
		}

		port, err := cmd.Flags().GetInt("port")
		if err == nil && port > 0 && port < 65535 {
			currentConfig.Port = port
		}
	}

	return &currentConfig
}

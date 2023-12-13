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
	Debug                    bool
	PIDFile                  string
	Port                     int
	Daemonize                bool
	PasswordSalt             string
	OTLPEndpoint             string
	InitialAdminUserPassword string
	CORSHosts                []string
	OtherSupernodes          []string
	TrustedProxies           []string
	HIBPAPIKey               string
	ServerName               string
	Supernode                bool
	Masquerade               bool
	WireguardTapAddress      string
	NodeIP                   string
	SupernodeZone            string
	strSessionSecret         string
	SessionSecret            []byte
	VTUNStartingAddress      string
	PostgresDSN              string
	postgresUser             string
	postgresPassword         string
	postgresHost             string
	postgresPort             int
	postgresDatabase         string
	MetricsNodeExporterHost  string
	MetricsPort              int
	Latitude                 string
	Longitude                string
	Gridsquare               string
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

	tmpConfig := Config{
		Debug:                    os.Getenv("DEBUG") != "",
		PIDFile:                  os.Getenv("PID_FILE"),
		Port:                     int(httpPort),
		Daemonize:                os.Getenv("NO_DAEMON") == "",
		PasswordSalt:             os.Getenv("PASSWORD_SALT"),
		OTLPEndpoint:             os.Getenv("OTLP_ENDPOINT"),
		InitialAdminUserPassword: os.Getenv("INIT_ADMIN_USER_PASSWORD"),
		HIBPAPIKey:               os.Getenv("HIBP_API_KEY"),
		ServerName:               os.Getenv("SERVER_NAME"),
		Supernode:                os.Getenv("SUPERNODE") != "",
		Masquerade:               os.Getenv("MASQUERADE") != "",
		WireguardTapAddress:      os.Getenv("WIREGUARD_TAP_ADDRESS"),
		NodeIP:                   os.Getenv("NODE_IP"),
		SupernodeZone:            os.Getenv("SUPERNODE_ZONE"),
		strSessionSecret:         os.Getenv("SESSION_SECRET"),
		VTUNStartingAddress:      os.Getenv("VTUN_STARTING_ADDRESS"),
		postgresUser:             os.Getenv("PG_USER"),
		postgresPassword:         os.Getenv("PG_PASSWORD"),
		postgresHost:             os.Getenv("PG_HOST"),
		postgresPort:             int(pgPort),
		postgresDatabase:         os.Getenv("PG_DATABASE"),
		MetricsNodeExporterHost:  os.Getenv("METRICS_NODE_EXPORTER_HOST"),
		MetricsPort:              int(metricsPort),
		Latitude:                 os.Getenv("SERVER_LAT"),
		Longitude:                os.Getenv("SERVER_LON"),
		Gridsquare:               os.Getenv("SERVER_GRIDSQUARE"),
	}

	if tmpConfig.VTUNStartingAddress == "" {
		tmpConfig.VTUNStartingAddress = "172.31.180.12"
	}

	if net.ParseIP(tmpConfig.VTUNStartingAddress) == nil {
		panic("VTUN starting address is not a valid IP address")
	}

	if tmpConfig.PIDFile == "" {
		tmpConfig.PIDFile = "/var/run/aredn-manager.pid"
	}

	if tmpConfig.Supernode && tmpConfig.SupernodeZone == "" {
		panic("Supernode zone not set")
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

		pidFile, err := cmd.Flags().GetString("pid-file")
		if err == nil && pidFile != "" {
			currentConfig.PIDFile = pidFile
		}

		port, err := cmd.Flags().GetInt("port")
		if err == nil && port > 0 && port < 65535 {
			currentConfig.Port = port
		}

		daemonize, err := cmd.Flags().GetBool("no-daemon")
		if err == nil && daemonize {
			currentConfig.Daemonize = false
		}
	}

	if currentConfig.Debug {
		fmt.Println(currentConfig.ToString())
	}

	return &currentConfig
}

// ToString returns a string representation of the configuration
func (config *Config) ToString() string {
	return "Debug: " + strconv.FormatBool(config.Debug) + "\n" +
		"PIDFile: " + config.PIDFile + "\n" +
		"Port: " + strconv.Itoa(config.Port) + "\n" +
		"Daemonize: " + strconv.FormatBool(config.Daemonize) + "\n" +
		"PasswordSalt: " + config.PasswordSalt + "\n" +
		"OTLPEndpoint: " + config.OTLPEndpoint + "\n" +
		"InitialAdminUserPassword: " + config.InitialAdminUserPassword + "\n" +
		"HIBPAPIKey: " + config.HIBPAPIKey + "\n" +
		"PostgresDSN: " + config.PostgresDSN + "\n" +
		"PostgresUser: " + config.postgresUser + "\n" +
		"PostgresPassword: " + config.postgresPassword + "\n" +
		"PostgresHost: " + config.postgresHost + "\n" +
		"PostgresPort: " + strconv.Itoa(config.postgresPort) + "\n" +
		"PostgresDatabase: " + config.postgresDatabase + "\n" +
		"CORSHosts: " + strings.Join(config.CORSHosts, ",") + "\n" +
		"TrustedProxies: " + strings.Join(config.TrustedProxies, ",") + "\n" +
		"MetricsPort: " + strconv.Itoa(config.MetricsPort) + "\n"
}

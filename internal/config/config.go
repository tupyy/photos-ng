package config

//go:generate go run github.com/ecordell/optgen -output zz_generated.configuration.go . Config
type Config struct {
	Database *Database `debugmap:"visible"`
	HttpPort int       `debugmap:"visible" default:"8080"`
	GrpcPort int       `debugmap:"visible" default:"9090"`

	DataRootFolder string `debugmap:"visible"`
	GinMode        string `debugmap:"visible"`
	Mode           string `debugmap:"visible" default:"dev"`
	StaticsFolder  string `debugmap:"visible"`

	// Log
	LogFormat      string         `debugmap:"visible"`
	LogLevel       string         `debugmap:"visible"`
	Authentication Authentication `debugmap:"visible"`
	Authorization  Authorization  `debugmap:"visible"`
}

//go:generate go run github.com/ecordell/optgen -output zz_generated.db_configuration.go . Database
type Database struct {
	URI                string `debugmap:"visible"`
	SSL                bool   `debugmap:"visible"`
	MaxOpenConnections int    `debugmap:"visible" default:"10"`
	Debug              bool   `debugmap:"visible" default:"false"`
}

//go:generate go run github.com/ecordell/optgen -output zz_generated.auth_configuration.go . Authentication
type Authentication struct {
	Enabled      bool   `debugmap:"visible"`
	WellknownURL string `debugmap:"visible"`
	ClientID     string `debugmap:"visible"`
	ClientSecret string `debugmap:"visible"`
}

type Authorization struct {
	Enabled      bool   `debugmap:"visible"`
	SpiceDBURL   string `debugmap:"visible" default:"localhost:50051"`
	PresharedKey string `debugmap:"visible" default:"dev-secret-key"`
}

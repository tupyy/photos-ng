package config

//go:generate go run github.com/ecordell/optgen -output zz_generated.configuration.go . Config
type Config struct {
	Database   *Database `debugmap:"visible"`
	ServerPort int       `debugmap:"visible" default:"8080"`
	GrpcPort   int       `debugmap:"visible" default:"9090"`

	DataRootFolder string `debugmap:"visible"`
	GinMode        string `debugmap:"visible"`
	Mode           string `debugmap:"visible" default:"dev"`
	StaticsFolder  string `debugmap:"visible"`

	// Log
	LogFormat string `debugmap:"visible"`
	LogLevel  string `debugmap:"visible"`
}

//go:generate go run github.com/ecordell/optgen -output zz_generated.db_configuration.go . Database
type Database struct {
	URI                string `debugmap:"visible"`
	SSL                bool   `debugmap:"visible"`
	MaxOpenConnections int    `debugmap:"visible" default:"10"`
	Debug              bool   `debugmap:"visible" default:"false"`
}

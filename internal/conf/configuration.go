package conf

import (
	"fmt"
	"log"
	"os"

	"github.com/gocql/gocql"
	"github.com/hiumesh/go-chat-server/internal/utils"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type APIConfiguration struct {
	Host string
	Port string `envconfig:"GO_SOCKET_PORT" default:"8080"`
}

func (c *APIConfiguration) Validate() error {
	return nil
}

type DBConfiguration struct {
	Host           string `envconfig:"GO_SOCKET_SCYLLA_HOST"`
	Keyspace       string `envconfig:"GO_SOCKET_SCYLLA_KEYSPACE"`
	MigrationsPath string `envconfig:"GO_SOCKET_SCYLLA_MIGRATIONS"`
	ClusterConfig  gocql.ClusterConfig
}

func (c *DBConfiguration) Validate() error {
	return nil
}

type REDISConfiguration struct {
	URL string `envconfig:"GO_SOCKET_REDIS_URL"`
}

func (c *REDISConfiguration) Validate() error {
	return nil
}

type CORSConfiguration struct {
	AllowedHeaders []string `json:"allowed_headers" split_words:"true"`
}

func (c *CORSConfiguration) AllAllowedHeaders(defaults []string) []string {
	set := make(map[string]bool)
	for _, header := range defaults {
		set[header] = true
	}

	var result []string
	result = append(result, defaults...)

	for _, header := range c.AllowedHeaders {
		if !set[header] {
			result = append(result, header)
		}

		set[header] = true
	}

	return result
}

type GlobalConfiguration struct {
	API   APIConfiguration
	DB    DBConfiguration
	REDIS REDISConfiguration
	CORS  CORSConfiguration `json:"cors"`
}

func generateServerName(purpose, location string) (string, string, error) {
	ipAddress, err := utils.GetIPAddress()
	if err != nil {
		log.Fatal(err)
	}

	serverName := fmt.Sprintf("%s-%s-%s", purpose, location, ipAddress)
	return serverName, ipAddress, nil
}

// func init() {
// 	location := os.Getenv("LOCATION")
// 	if location == "" {
// 		log.Fatal("LOCATION environment not found.")
// 	}
// 	generateServerName("websocket-server", location)
// }

func loadEnvironment(filename string) error {
	var err error
	if filename != "" {
		err = godotenv.Overload(filename)
	} else {
		err = godotenv.Load()
		if os.IsNotExist(err) {
			return nil
		}
	}
	return err
}

func LoadGlobal(filename string) (*GlobalConfiguration, error) {
	if err := loadEnvironment(filename); err != nil {
		return nil, err
	}

	config := new(GlobalConfiguration)
	if err := envconfig.Process("gosocket", config); err != nil {
		return nil, err
	}

	print(config)

	if err := config.Validate(); err != nil {
		return nil, err
	}
	return config, nil
}

func (c *GlobalConfiguration) Validate() error {
	validatables := []interface {
		Validate() error
	}{
		&c.API,
		&c.DB,
		&c.REDIS,
	}

	for _, validatable := range validatables {
		if err := validatable.Validate(); err != nil {
			return err
		}
	}

	return nil
}

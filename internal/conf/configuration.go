package conf

import (
	"os"

	"github.com/gocql/gocql"
	"github.com/hiumesh/go-chat-server/internal/utils"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type ServerConfiguration struct {
	Id                   string
	MaxPerUserConnection string `envconfig:"GO_SOCKET_MAX_PER_USER_CONNECTION" default:"2"`
}

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

type CookieConfiguration struct {
	Key      string `json:"key"`
	Domain   string `json:"domain"`
	Duration int    `json:"duration"`
}

type JWTConfiguration struct {
	Secret string `json:"secret" required:"true"`
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
	SERVER ServerConfiguration
	API    APIConfiguration
	DB     DBConfiguration
	REDIS  REDISConfiguration
	CORS   CORSConfiguration   `json:"cors"`
	JWT    JWTConfiguration    `json:"jwt"`
	COOKIE CookieConfiguration `json:"cookies"`
}

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

	config.SERVER.Id = utils.GenerateUniqueServerId()

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

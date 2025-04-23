package orm

import (
	"fmt"
	"os"
	"strconv"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Migrator = func(*gorm.DB) error

type ORM struct {
	db *gorm.DB
}

func (o *ORM) DB() *gorm.DB {
	return o.db
}
func (o *ORM) Migrate(migrators []Migrator) error {
	for _, migrator := range migrators {
		err := migrator(o.db)
		if err != nil {
			return err
		}
	}

	return nil
}

func New(
	host string,
	port uint,
	user string,
	password string,
	dbname string,
	sslMode string,
	timeZone string,
) (*ORM, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		host, port, user, password, dbname, sslMode, timeZone,
	)

	db, err := gorm.Open(postgres.Open(dsn))
	if err != nil {
		return nil, err
	}

	err = db.Exec("create schema if not exists dating").Error
	if err != nil {
		return nil, err
	}

	return &ORM{db: db}, nil
}

func NewFromEnvironments() (*ORM, error) {
	host := getEnvOrDefault("DB_HOST", "localhost")
	portString := getEnvOrDefault("DB_PORT", "5432")
	user := getEnvOrDefault("DB_USERNAME", "dating")
	password := getEnvOrDefault("DB_PASSWORD", "dating")
	dbname := getEnvOrDefault("DB_NAME", "dating")

	port, err := strconv.ParseUint(portString, 10, 0)
	if err != nil {
		return nil, err
	}

	orm, err := New(host, uint(port), user, password, dbname, "disable", "Europe/Moscow")
	if err != nil {
		return nil, err
	}

	return orm, nil
}

func getEnv(key string) (string, error) {
	value, ok := os.LookupEnv(key)
	if !ok {
		return "", fmt.Errorf("not found env variable by %s name", key)
	}

	return value, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value, err := getEnv(key); err == nil {
		return value
	}

	return defaultValue
}

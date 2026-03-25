package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

var (
	secretsPath = "/vault/secrets/"
	defaultPort = 8080
)

// OAuthCfg хранит параметры OAuth2-провайдера.
type OAuthCfg struct {
	Provider     string
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

// S3Cfg хранит параметры S3-совместимого хранилища.
type S3Cfg struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	Bucket          string
	BaseURL         string
	Region          string
}

// SMTPCfg хранит параметры SMTP для отправки email.
type SMTPCfg struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

// MultiSMTPCfg хранит несколько SMTP конфигураций.
type MultiSMTPCfg struct {
	Primary   SMTPCfg
	Secondary SMTPCfg
	Tertiary  SMTPCfg
}

// Config хранит параметры приложения.
type Config struct {
	AppURL             string
	Environment        string
	LogLevel           string
	Port               int
	DBConnectionString string
	JWTSecret          string
	OAuth              OAuthCfg
	S3                 S3Cfg
	SMTP               MultiSMTPCfg
}

// readValueFromFileOrEnv сначала пытается прочитать секрет из файла,
// который Vault Agent записывает при деплое в /vault/secrets/CRED_<NAME>.
// Если файл недоступен (локальная разработка) читает из переменной окружения.
func readValueFromFileOrEnv(valueName string) string {
	value, err := os.ReadFile(secretsPath + "CRED_" + valueName)
	if err != nil {
		log.Printf("Unable to read value from file %s: %s, trying environment variable", valueName, err)

		return os.Getenv(valueName)
	}

	return string(value)
}

func mustReadValueFromFileOrEnv(valueName string) string {
	value := readValueFromFileOrEnv(valueName)
	if value == "" {
		log.Printf("Unable to read required value %s", valueName)
		os.Exit(1)
	}

	return value
}

func readValueAsInt(valueName string, defaultVal int) int {
	valueStr := readValueFromFileOrEnv(valueName)
	if valueStr == "" {
		return defaultVal
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		log.Printf("Unable to convert value %s to int: %v, using default value %d", valueName, err, defaultVal)

		return defaultVal
	}

	return value
}

func makeDbConnectionString() string {
	dbUsername := mustReadValueFromFileOrEnv("DB_USERNAME")
	dbPassword := mustReadValueFromFileOrEnv("DB_PASSWORD")
	dbName := mustReadValueFromFileOrEnv("DB_NAME")
	dbHost := mustReadValueFromFileOrEnv("DB_HOST")
	dbPort := readValueAsInt("DB_PORT", 5432)

	return "postgres://" + dbUsername + ":" + dbPassword +
		"@" + dbHost + ":" + strconv.Itoa(dbPort) + "/" + dbName + "?sslmode=disable"
}

// MustLoad загружает конфигурацию из .env-файла и переменных окружения.
func MustLoad() *Config {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	port := readValueAsInt("APP_PORT", defaultPort)

	appURL := readValueFromFileOrEnv("APP_URL")
	if appURL == "" {
		appURL = "localhost:" + strconv.Itoa(port)
	}

	jwtSecret := readValueFromFileOrEnv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "default-jwt-secret-for-development"
	}

	return &Config{
		AppURL:             appURL,
		Environment:        mustReadValueFromFileOrEnv("APP_ENV"),
		LogLevel:           readValueFromFileOrEnv("LOG_LEVEL"),
		Port:               port,
		DBConnectionString: makeDbConnectionString(),
		JWTSecret:          jwtSecret,
		OAuth: OAuthCfg{
			Provider:     mustReadValueFromFileOrEnv("OAUTH_PROVIDER"),
			ClientID:     mustReadValueFromFileOrEnv("OAUTH_CLIENT_ID"),
			ClientSecret: mustReadValueFromFileOrEnv("OAUTH_CLIENT_SECRET"),
			RedirectURL:  mustReadValueFromFileOrEnv("OAUTH_REDIRECT_URL"),
		},
		S3: S3Cfg{
			Endpoint:        readValueFromFileOrEnv("S3_ENDPOINT"),
			AccessKeyID:     readValueFromFileOrEnv("S3_ACCESS_KEY_ID"),
			SecretAccessKey: readValueFromFileOrEnv("S3_SECRET_ACCESS_KEY"),
			Bucket:          readValueFromFileOrEnv("S3_BUCKET"),
			BaseURL:         readValueFromFileOrEnv("S3_BASE_URL"),
			Region:          readValueFromFileOrEnv("S3_REGION"),
		},
		SMTP: MultiSMTPCfg{
			Primary: SMTPCfg{
				Host:     readValueFromFileOrEnv("SMTP_HOST"),
				Port:     readValueAsInt("SMTP_PORT", 587),
				Username: readValueFromFileOrEnv("SMTP_USERNAME"),
				Password: readValueFromFileOrEnv("SMTP_PASSWORD"),
				From:     readValueFromFileOrEnv("SMTP_FROM"),
			},
			Secondary: SMTPCfg{
				Host:     readValueFromFileOrEnv("SMTP_HOST_2"),
				Port:     readValueAsInt("SMTP_PORT_2", 587),
				Username: readValueFromFileOrEnv("SMTP_USERNAME_2"),
				Password: readValueFromFileOrEnv("SMTP_PASSWORD_2"),
				From:     readValueFromFileOrEnv("SMTP_FROM_2"),
			},
			Tertiary: SMTPCfg{
				Host:     readValueFromFileOrEnv("SMTP_HOST_3"),
				Port:     readValueAsInt("SMTP_PORT_3", 587),
				Username: readValueFromFileOrEnv("SMTP_USERNAME_3"),
				Password: readValueFromFileOrEnv("SMTP_PASSWORD_3"),
				From:     readValueFromFileOrEnv("SMTP_FROM_3"),
			},
		},
	}
}

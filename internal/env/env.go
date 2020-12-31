package env

// Bag holds all the config envvars
type Bag struct {
	ServerHost            string `envconfig:"SERVER_HOST"`
	ServerPort            string `envconfig:"SERVER_PORT" default:"8080"`
	ServerShutdownTimeout int    `envconfig:"SERVER_SHUTDOWN_TIMEOUT" default:"5"`
	HTTPTraceHeader       string `envconfig:"HTTP_TRACE_HEADER"`
	SMTPAddr              string `envconfig:"SMTP_ADDR" required:"true"`
	LogLevel              string `envconfig:"LOG_LEVEL" default:"info"`
}

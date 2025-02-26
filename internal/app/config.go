package app

type AppConfig struct {
	RAS_HOST      string `env:"RAS_HOST" envDefault:"localhost"`
	RAS_PORT      string `env:"RAS_PORT" envDefault:"1545"`
	RAS_VERSION   string `env:"RAS_VERSION" envDefault:"3.0"`
	CLS_USER      string `env:"CLS_USER"`
	CLS_PASS      string `env:"CLS_PASS"`
	AGNT_USER     string `env:"AGNT_USER"`
	AGNT_PASS     string `env:"AGNT_PASS"`
	MODE          string `env:"MODE" envDefault:"pull"`
	PUSH_INTERVAL int    `env:"PUSH_INTERVAL" envDefault:"500"`
	PUSH_HOST     string `env:"PUSH_HOST" envDefault:"localhost"`
	PUSH_PORT     string `env:"PUSH_PORT" envDefault:"9091"`
	PULL_EXPOSE   string `env:"PULL_EXPOSE" envDefault:"9096"`
}

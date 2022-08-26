package loggercheck

type Option func(*loggercheck)

func WithDisable(disable []string) Option {
	return func(l *loggercheck) {
		l.disable = loggerCheckersFlag{stringSet: newStringSet(disable...)}
	}
}

func WithConfig(cfg *Config) Option {
	return func(l *loggercheck) {
		l.config.cfg = cfg
	}
}

func WithDisableFlags(disable bool) Option {
	return func(l *loggercheck) {
		l.disableFlags = disable
	}
}

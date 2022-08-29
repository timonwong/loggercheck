package loggercheck

type Option func(*loggercheck)

func WithConfig(cfg *Config) Option {
	return func(l *loggercheck) {
		l.cfg = cfg
	}
}

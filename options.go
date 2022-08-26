package loggercheck

type Option func(*loggercheck)

func WithDisable(disable []string) Option {
	return func(l *loggercheck) {
		l.disable = loggerCheckersFlag{stringSet: newStringSet(disable...)}
	}
}

func WithCustomLogger(name, packageImport string, funcs []string) Option {
	return func(l *loggercheck) {
		addLogger(name, packageImport, funcs)
	}
}

func WithDisableFlags(disable bool) Option {
	return func(l *loggercheck) {
		l.disableFlags = disable
	}
}

package noprintflike

import "github.com/go-logr/logr"

func ExamplePrintfLike() {
	log := logr.Discard()

	// no formats
	log.Info("hello")

	// invalid formats
	const InvalidFormat = "hello %% %1 %2 %3"
	log.Info(InvalidFormat)
	log.Info("hello %[s")
	log.Info("hello %[-1]s")
	log.Info("hello %[0]s")
	log.Info("hello %")
	log.Info("hello %#.d")
	log.Info("%.3[1f")
	log.Info("d%")
	log.Info("%.3")
	log.Info("%#[1].3")
	log.Info("%[3]*.[2*[1]f", "intKey", 1)

	log.Info("%[3]*s x") // want `logging message should not use format specifier ".+"`
	log.Info("%[3]d x")  // want `logging message should not use format specifier ".+"`

	log.Info("% 8s")                        // want `logging message should not use format specifier "% 8s"`
	log.Info("hello %s", "intKey", 1)       // want `logging message should not use format specifier "%s"`
	log.Info("%.3[1]f", "intKey", 1)        // want `logging message should not use format specifier ".+"`
	log.Info("%[3]*.[2]*[1]f", "intKey", 1) // want `logging message should not use format specifier ".+"`
	const ValidFormat = "hello %#v %32d %f %d %g %% %s %9.2f %w %T %[1]d"
	log.Info(ValidFormat, "intKey", 1) // want `logging message should not use format specifier "%#v"`
}

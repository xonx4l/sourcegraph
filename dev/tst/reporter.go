package tst

import "fmt"

type Reporter interface {
	Writef(format string, args ...any) (int, error)
	Writeln(v string) (int, error)
	Prefix(p string)
}

type ConsoleReporter struct {
	prefix string
}
type NoopReporter struct{}

func (r *ConsoleReporter) Prefix(p string) {
	r.prefix = p
}

func (r *ConsoleReporter) Writef(format string, args ...any) (int, error) {
	fmt.Printf(r.prefix)
	return fmt.Printf(format, args...)
}

func (r *ConsoleReporter) Writeln(v string) (int, error) {
	return fmt.Println(v)
}

func (r NoopReporter) Prefix(p string) {}

func (r NoopReporter) Writef(format string, args ...any) (int, error) {
	return 0, nil
}

func (r NoopReporter) Writeln(v string) (int, error) {
	return 0, nil
}

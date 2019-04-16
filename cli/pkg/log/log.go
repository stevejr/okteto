package log

import (
	"fmt"
	"os"
	"runtime"

	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
)

var (
	redString = color.New(color.FgHiRed).SprintfFunc()

	greenString = color.New(color.FgGreen).SprintfFunc()

	yellowString = color.New(color.FgHiYellow).SprintfFunc()

	blueString = color.New(color.FgHiBlue).SprintfFunc()

	errorSymbol = color.New(color.BgHiRed, color.FgBlack).Sprint(" x ")

	successSymbol = color.New(color.BgGreen, color.FgBlack).Sprint(" ✓ ")

	informationSymbol = color.New(color.BgHiBlue, color.FgBlack).Sprint(" i ")
)

type logger struct {
	out *logrus.Logger
}

var log = &logger{
	out: logrus.New(),
}

func init() {
	if runtime.GOOS == "windows" {
		successSymbol = color.New(color.BgGreen, color.FgBlack).Sprint(" + ")
	}
}

// Init configures the logger for the package to use.
func Init(level logrus.Level) {
	log.out.SetOutput(os.Stdout)
	log.out.SetLevel(level)
}

// SetLevel sets the level of the main logger
func SetLevel(level string) {
	l, err := logrus.ParseLevel(level)
	if err == nil {
		log.out.SetLevel(l)
	}
}

// Debug writes a debug-level log
func Debug(args ...interface{}) {
	log.out.Debug(args...)
}

// Debugf writes a debug-level log with a format
func Debugf(format string, args ...interface{}) {
	log.out.Debugf(format, args...)
}

// Info writes a info-level log
func Info(args ...interface{}) {
	log.out.Info(args...)
}

// Infof writes a info-level log with a format
func Infof(format string, args ...interface{}) {
	log.out.Infof(format, args...)
}

// Fatal writes a fatal-level log
func Fatal(args ...interface{}) {
	log.out.Fatal(args...)
}

// Error writes a error-level log
func Error(args ...interface{}) {
	log.out.Error(args...)
}

// Errorf writes a error-level log with a format
func Errorf(format string, args ...interface{}) {
	log.out.Errorf(format, args...)
}

// Yellow writes a line in yellow
func Yellow(format string, args ...interface{}) {
	fmt.Fprintln(color.Output, yellowString(format, args...))
}

// Green writes a line in green
func Green(format string, args ...interface{}) {
	fmt.Fprintln(color.Output, greenString(format, args...))
}

// Success prints a message with the success symbol first, and the text in green
func Success(format string, args ...interface{}) {
	fmt.Fprintf(color.Output, "%s %s\n", successSymbol, greenString(format, args...))
}

// Information prints a message with the information symbol first, and the text in blue
func Information(format string, args ...interface{}) {
	fmt.Fprintf(color.Output, "%s %s\n", informationSymbol, blueString(format, args...))
}

// Fail prints a message with the error symbol first, and the text in red
func Fail(format string, args ...interface{}) {
	fmt.Fprintf(color.Output, "%s %s\n", errorSymbol, redString(format, args...))
}
package logger

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
)

// Loggable -
type Loggable interface {
	LogFields() logrus.Fields
}

var logger = newBCDLogger()

// BCDLogger -
type BCDLogger struct {
	*logrus.Logger
}

func newBCDLogger() *BCDLogger {
	l := &BCDLogger{
		Logger: logrus.New(),
	}

	formatter := new(logrus.TextFormatter)
	switch os.Getenv("BCD_ENV") {
	case "development":
		formatter.FullTimestamp = true
		formatter.TimestampFormat = "2006-01-02 15:04:05"
	default:
		formatter.DisableTimestamp = true
	}

	l.SetFormatter(formatter)
	l.SetOutput(os.Stdout)
	l.SetLevel(logrus.InfoLevel)
	return l
}

// Error -
func Error(err error) {
	logger.Error(err)
}

// Errorf -
func Errorf(format string, args ...interface{}) {
	logger.Errorf(format, args...)
}

// Info -
func Info(format string, args ...interface{}) {
	logger.Infof(format, args...)
}

// Warning -
func Warning(format string, args ...interface{}) {
	logger.Warnf(format, args...)
}

// Fatal -
func Fatal(err error) {
	logger.Fatal(err)
	os.Exit(1)
}

// Log -
func Log(text string) {
	logger.Print(text)
}

// Logf -
func Logf(format string, args ...interface{}) {
	logger.Printf(format, args...)
}

// JSON - pretty json log
func JSON(data string) {
	var pretty bytes.Buffer
	if err := json.Indent(&pretty, []byte(data), "", "  "); err != nil {
		Error(err)
	} else {
		Info(pretty.String())
	}
}

// InterfaceToJSON - pretty json log
func InterfaceToJSON(data interface{}) {
	if result, err := json.MarshalIndent(data, "", "  "); err != nil {
		Error(err)
	} else {
		Info(string(result))
	}
}

// Debug -
func Debug(values ...interface{}) {
	if len(values) == 0 {
		return
	}
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		fmt.Printf("[DEBUG] can't capture stack for dump\n")
		return
	}
	targetFile, openErr := os.Open(file)
	if openErr != nil {
		fmt.Printf("[DEBUG] can't open file: %v\n", file)
		return
	}
	defer func() {
		err := targetFile.Close()
		if err != nil {
			fmt.Printf("[DEBUG] can't close file: %v; err: %v\n", file, err)
		}
	}()

	scanner := bufio.NewScanner(targetFile)
	lineCnt := 0
	targetLine := ""
	for scanner.Scan() {
		lineCnt++
		if lineCnt != line {
			continue
		}
		fileLine := strings.Trim(scanner.Text(), ` 	`)
		dumpStartIdx := strings.Index(fileLine, "Debug(")
		dumpEndIdx := strings.LastIndex(fileLine, ")")
		if dumpStartIdx < 0 || dumpEndIdx < 0 {
			fmt.Printf("[DEBUG] target line is invalid. Debug should start with `Debug(` and end with `)`: %v\n", fileLine)
			return
		}
		targetLine = fileLine[dumpStartIdx+5 : dumpEndIdx]
		break
	}
	dumpVariables := strings.Split(targetLine, ", ")
	if len(dumpVariables) != len(values) {
		buff := &bytes.Buffer{}
		_, _ = fmt.Fprintf(buff, "[DEBUG] %v:%v: ", file, line)
		_, _ = fmt.Fprintf(buff, "%v: ", targetLine)
		for idx, val := range values {
			_, _ = fmt.Fprintf(buff, "`%+v`", val)
			if idx < len(values)-1 {
				_, _ = fmt.Fprintf(buff, "; ")
			}
		}
		_, _ = fmt.Fprintf(buff, "\n")
		fmt.Print(buff.String())
		return
	}

	buff := &bytes.Buffer{}
	_, _ = fmt.Fprintf(buff, "[DEBUG] %v:%v: ", file, line)
	for idx, variable := range dumpVariables {
		isStringLiteral := strings.HasPrefix(variable, `"`) && strings.HasSuffix(variable, `"`)
		isStringLiteral = isStringLiteral || strings.HasPrefix(variable, "`") && strings.HasSuffix(variable, "`")
		if isStringLiteral {
			_, _ = fmt.Fprintf(buff, "%v", variable[1:len(variable)-1])
		} else {
			_, _ = fmt.Fprintf(buff, "%v: `%+v`", variable, values[idx])
			if idx < len(values)-1 {
				_, _ = fmt.Fprintf(buff, "; ")
			}
		}
	}
	_, _ = fmt.Fprintf(buff, "\n")
	fmt.Print(buff.String())
}

// Question -
func Question(format string, v ...interface{}) {
	blue := color.New(color.FgMagenta).SprintFunc()
	fmt.Printf("[%s] %s", blue("?"), fmt.Sprintf(format, v...))
}

// With -
func With(entry Loggable) *logrus.Entry {
	return logger.WithFields(entry.LogFields())
}

// WithNetwork -
func WithNetwork(network types.Network) *logrus.Entry {
	return logger.WithField("network", network.String())
}

// WithField -
func WithField(name string, value interface{}) *logrus.Entry {
	return logger.WithFields(logrus.Fields{
		name: value,
	})
}

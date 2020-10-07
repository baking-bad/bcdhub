package logger

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/fatih/color"
	"github.com/pkg/errors"
)

// Error -
func Error(err error) {
	red := color.New(color.FgRed).SprintFunc()
	log.Printf("[%s] %s", red("Error"), err)
}

// Errorf -
func Errorf(format string, v ...interface{}) {
	red := color.New(color.FgRed).SprintFunc()
	log.Printf("[%s] %s", red("Error"), errors.Errorf(format, v...))
}

// Info -
func Info(format string, v ...interface{}) {
	blue := color.New(color.FgBlue).SprintFunc()
	log.Printf("[%s] %s", blue("Info"), fmt.Sprintf(format, v...))
}

// Success -
func Success(format string, v ...interface{}) {
	green := color.New(color.FgGreen).SprintFunc()
	log.Printf("[%s] %s", green("Success"), fmt.Sprintf(format, v...))
}

// Warning -
func Warning(format string, v ...interface{}) {
	yellow := color.New(color.FgYellow).SprintFunc()
	log.Printf("[%s] %s", yellow("Warning"), fmt.Sprintf(format, v...))
}

// Fatal -
func Fatal(err error) {
	red := color.New(color.FgRed).SprintFunc()
	log.Printf("[%s] %s", red("FATAL"), err)
	os.Exit(1)
}

// Log -
func Log(text string) {
	log.Print(text)
}

// Logf -
func Logf(format string, v ...interface{}) {
	log.Printf(format, v...)
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
	log.Printf("[%s] %s", blue("?"), fmt.Sprintf(format, v...))
}

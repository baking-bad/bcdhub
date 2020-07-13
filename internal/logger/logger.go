package logger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"

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

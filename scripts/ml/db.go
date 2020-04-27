package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/baking-bad/bcdhub/internal/logger"
	"golang.org/x/crypto/ssh/terminal"
)

func askQuestion(question string) (string, error) {
	logger.Warning(question)

	reader := bufio.NewReader(os.Stdin)
	text, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.Replace(text, "\n", "", -1), nil
}

func askPassword(question string) (string, error) {
	logger.Warning(question)

	text, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}
	return string(text), nil
}

// user=root dbname=bcd password=root
func askDatabaseConnectionString(connString string) (string, error) {
	dbName, err := askQuestion("Enter DB name (default: bcd)")
	if err != nil {
		return "", err
	}
	if dbName == "" {
		dbName = "bcd"
	}
	dbUser, err := askQuestion("Enter DB user (default: root)")
	if err != nil {
		return "", err
	}
	if dbUser == "" {
		dbUser = "root"
	}

	dbPass, err := askPassword("Enter DB password")
	if err != nil {
		return "", err
	}
	connString = fmt.Sprintf("%s user=%s dbname=%s password=%s", connString, dbUser, dbName, dbPass)
	return connString, nil
}

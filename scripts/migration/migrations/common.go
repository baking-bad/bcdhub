package migrations

import (
	"bufio"
	"os"
	"strings"

	"github.com/baking-bad/bcdhub/internal/logger"
)

func ask(question string) (string, error) {
	logger.Question(question)
	reader := bufio.NewReader(os.Stdin)
	text, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.ReplaceAll(text, "\n", ""), nil
}

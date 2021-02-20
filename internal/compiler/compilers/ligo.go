package compilers

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/baking-bad/bcdhub/internal/bcd/contract/language"
)

type ligo struct{}

// Language -
func (c *ligo) Language() string {
	return language.LangLigo
}

// Compile -
func (c *ligo) Compile(path string) (*Data, error) {
	if err := c.checkExtension(path); err != nil {
		return nil, err
	}

	cmd := exec.Command(LigoPath, "compile-contract", "--michelson-format=json", path, "main")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	return &Data{
		Script:   string(out),
		Language: c.Language(),
	}, nil
}

func (c *ligo) checkExtension(path string) error {
	ext := filepath.Ext(path)

	entrypoints := map[string]string{
		".ligo":   "function main",
		".religo": "let main",
		".mligo":  "let main",
	}

	entrypoint, ok := entrypoints[ext]
	if !ok {
		return fmt.Errorf("invalid file extension %v", path)
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), entrypoint) {
			return nil
		}
	}

	return fmt.Errorf("entrypoint not found")
}

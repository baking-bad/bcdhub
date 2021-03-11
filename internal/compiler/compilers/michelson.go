package compilers

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/baking-bad/bcdhub/internal/bcd/contract/language"
)

type michelson struct{}

// Language -
func (c *michelson) Language() string {
	return language.LangMichelson
}

// Compile -
func (c *michelson) Compile(path string) (*Data, error) {
	if err := c.checkExtension(path); err != nil {
		return nil, err
	}

	cmd := exec.Command(MichelsonPath, "-mode", "mockup", "convert", "script", path, "from", "michelson", "to", "json")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%v %v", string(out), err)
	}

	return &Data{
		Script:   string(out),
		Language: c.Language(),
	}, nil
}

func (c *michelson) checkExtension(path string) error {
	if ext := filepath.Ext(path); ext != ".tz" {
		return fmt.Errorf("invalid file extension %v", path)
	}

	return nil
}

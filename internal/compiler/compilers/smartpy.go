package compilers

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/baking-bad/bcdhub/internal/bcd/contract/language"
)

type smartpy struct{}

// Language -
func (c *smartpy) Language() string {
	return language.LangSmartPy
}

// Compile -
func (c *smartpy) Compile(path string) (*Data, error) {
	if err := c.checkExtension(path); err != nil {
		return nil, err
	}

	tempDir, err := ioutil.TempDir("/tmp", "smartpy-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tempDir)

	entrypoint, err := c.findEntrypoint(path)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(SmartpyPath, "compile", path, entrypoint, tempDir)
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	resultPath := fmt.Sprintf("%s/%s_compiled.json", tempDir, filenameByPath(path))

	data, err := ioutil.ReadFile(resultPath)
	if err != nil {
		return nil, err
	}

	return &Data{
		Script:   string(data),
		Language: c.Language(),
	}, nil
}

func (c *smartpy) checkExtension(path string) error {
	if ext := filepath.Ext(path); ext != ".py" {
		return fmt.Errorf("invalid file extension %v", path)
	}

	return nil
}

func (c *smartpy) findEntrypoint(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	smartPyRe := regexp.MustCompile(`^class (\w+)\(sp.Contract\):$`)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if res := checkRegexp(smartPyRe, scanner.Text()); res != "" {
			return fmt.Sprintf("%s()", res), nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", fmt.Errorf("not found")
}

func checkRegexp(re *regexp.Regexp, line string) string {
	m := re.FindAllStringSubmatch(line, -1)
	if m != nil && len(m[0]) == 2 {
		return m[0][1]
	}

	return ""
}

func filenameByPath(path string) string {
	base := filepath.Base(path)
	return strings.TrimSuffix(base, filepath.Ext(base))
}

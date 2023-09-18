package dotenv

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

type Environment map[string]string

const COMMENT = '#'

func Load(file *os.File) *Environment {
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	oldEnv := make(Environment)
	for scanner.Scan() {
		line := scanner.Text()

		if len(line) == 0 || line[0] == COMMENT {
			continue
		}

		index := strings.Index(line, "=")
		if index == -1 {
			continue
		}
		key := strings.TrimSpace(line[:index])
		value := strings.TrimSpace(line[index+1:])

		oldEnv[key] = os.Getenv(key)
		os.Setenv(key, value)
	}

	file.Close()

	return &oldEnv
}

func (env *Environment) Restore() {
	for key, value := range *env {
		if len(value) == 0 {
			os.Unsetenv(key)
		} else {
			os.Setenv(key, value)
		}
	}
}

func GetInt(key string) int {
	value, err := strconv.ParseInt(os.Getenv(key), 10, 32)
	if err != nil {
		return 0
	}
	return int(value)
}

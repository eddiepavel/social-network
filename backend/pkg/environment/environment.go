package environment

import (
	"bufio"
	"errors"
	"os"
	"strings"
)

func SetEnv(filePath string) error {
	file, err := os.Open(filePath)

	if err != nil {
		return errors.New("env file not found")
	}

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.Split(scanner.Text(), "=")

		if len(line) != 2 {
			return errors.New("env file corrupted")
		}

		os.Setenv(line[0], line[1])
	}

	return nil
}

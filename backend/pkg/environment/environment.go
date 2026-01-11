package environment

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

func SetEnv(filePath string) error {
	file, err := os.Open(filePath)

	if err != nil {
		return errors.New("env file not found")
	}

	scanner := bufio.NewScanner(file)

	var envKeys []string
	compareKeys := []string{"PORT", "DATABASE_NAME", "PRODUCTION", "APP_URL", "SECRET_SIGN"}

	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.Split(scanner.Text(), "=")

		switch {
		case len(line) != 2:
			return fmt.Errorf("env file corrupted")
		case len(line[0]) <= 0:
			return fmt.Errorf("missing key in line %d", lineNum)
		case len(line[1]) <= 0:
			return fmt.Errorf("missing value in %s", line[0])
		}

		envKeys = append(envKeys, line[0])

		os.Setenv(line[0], line[1])
	}

	missing := findMissing(compareKeys, envKeys)
	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %v", missing)
	}

	return nil
}

func findMissing(required []string, actual []string) []string {
	actualMap := make(map[string]bool)
	for _, key := range actual {
		actualMap[key] = true
	}

	var missing []string
	for _, key := range required {
		if !actualMap[key] {
			missing = append(missing, key)
		}
	}

	return missing
}

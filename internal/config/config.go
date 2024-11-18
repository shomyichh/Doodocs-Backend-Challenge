package config

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

func Init() {
	err := loadEnvFile("../.env")

	if err != nil {
		fmt.Println(err)
	}

	if os.Getenv("TMPDIR") == "" {
		os.Setenv("TMPDIR", "/default/temp")
	} else {
		os.Setenv("TMPDIR", os.Getenv("TMPDIR"))
	}
	requiredVars := []string{"SMTP_HOST", "SMTP_PORT", "SMTP_USER", "SMTP_PASS"}
	for _, v := range requiredVars {
		if os.Getenv(v) == "" {
			log.Fatalf("Environment variable %s is not set", v)
		}
	}
}

func Get(key string) string {
	return os.Getenv(key)
}

func loadEnvFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("could not open .env file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" || line[0] == '#' {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			os.Setenv(key, value)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading .env file: %v", err)
	}

	return nil
}

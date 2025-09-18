package config

import (
	"bufio"
	"os"
	"strings"
)

// LoadEnvFile загружает переменные окружения из .env файла
func LoadEnvFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		// Файл .env не обязателен, просто возвращаем nil
		return nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// Пропускаем пустые строки и комментарии
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Парсим KEY=VALUE
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Удаляем кавычки если есть
		if len(value) >= 2 {
			if (strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`)) ||
				(strings.HasPrefix(value, `'`) && strings.HasSuffix(value, `'`)) {
				value = value[1 : len(value)-1]
			}
		}

		// Устанавливаем переменную окружения если она еще не установлена
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}

	return scanner.Err()
}
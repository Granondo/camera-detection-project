package config

import (
	"bufio"
	"log"
	"os"
	"strings"
)

// LoadEnvFile загружает переменные окружения из .env файла
func LoadEnvFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		log.Printf("📄 .env файл '%s' не найден, используем системные переменные", filename)
		return nil
	}
	defer file.Close()
	
	log.Printf("📄 Загружаем переменные из файла: %s", filename)

	scanner := bufio.NewScanner(file)
	count := 0
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// Пропускаем пустые строки и комментарии
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Парсим KEY=VALUE
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			log.Printf("⚠️ Пропускаем неверный формат строки: %s", line)
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
			// Маскируем пароли в логах
			displayValue := value
			if strings.Contains(strings.ToLower(key), "password") {
				displayValue = maskValue(value)
			}
			log.Printf("📝 Загружен %s = %s", key, displayValue)
			count++
		} else {
			log.Printf("⏭️ Пропущен %s (уже установлен в системе)", key)
		}
	}
	
	log.Printf("✅ Загружено %d переменных из .env файла", count)
	return scanner.Err()
}

// maskValue маскирует значение для безопасного вывода
func maskValue(value string) string {
	if value == "" {
		return "(пустое)"
	}
	if len(value) <= 3 {
		return "***"
	}
	return value[:1] + "***" + value[len(value)-1:]
}
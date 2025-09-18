package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"camera-detection-project/internal/camera"
	"camera-detection-project/internal/config"
)

func main() {
	log.Println("Запуск Camera Detection Project...")

	// Загружаем конфигурацию
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	// Создаем контекст для graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Создаем камеру (используем простую версию через OpenCV)
	cam := camera.NewSimple(cfg.Camera)

	// WaitGroup для координации горутин
	var wg sync.WaitGroup

	// Запускаем камеру в отдельной горутине
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := cam.Start(ctx); err != nil {
			log.Printf("Ошибка работы камеры: %v", err)
		}
	}()

	// Обрабатываем сигналы для graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Ждем сигнал завершения
	<-sigChan
	log.Println("Получен сигнал завершения, останавливаем приложение...")

	// Отменяем контекст
	cancel()

	// Ждем завершения всех горутин
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	// Ждем завершения с таймаутом
	select {
	case <-done:
		log.Println("Приложение успешно остановлено")
	case <-time.After(10 * time.Second):
		log.Println("Таймаут при остановке приложения")
	}
}
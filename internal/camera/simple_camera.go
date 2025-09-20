package camera

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"time"

	"camera-detection-project/internal/config"
	"gocv.io/x/gocv"
)

// SimpleCamera использует OpenCV для получения кадров через RTSP
type SimpleCamera struct {
	config     config.CameraConfig
	videoCapture *gocv.VideoCapture
	frameNum   int
	isRunning  bool
}

// NewSimple создает простой экземпляр камеры
func NewSimple(cfg config.CameraConfig) *SimpleCamera {
	return &SimpleCamera{
		config: cfg,
	}
}

// Start запускает получение видеопотока через OpenCV
func (c *SimpleCamera) Start(ctx context.Context) error {
	log.Printf("=== 📹 СТАРТ КАМЕРЫ ===")
	log.Printf("🔗 RTSP URL: %s", c.maskURL(c.config.RTSPUrl))

	// Проверяем FFmpeg
	log.Println("🔍 Проверяем FFmpeg...")
	if err := c.checkFFmpeg(); err != nil {
		log.Printf("⚠️ FFmpeg недоступен: %v", err)
	} else {
		log.Println("✅ FFmpeg найден")
	}

	// Строим RTSP URL
	rtspURL := c.buildRTSPURL()
	log.Printf("🎯 Подключаемся к: %s", c.maskURL(rtspURL))

	// Создаем VideoCapture
	webcam, err := gocv.OpenVideoCapture(rtspURL)
	if err != nil {
		return fmt.Errorf("❌ ошибка открытия видеопотока: %v", err)
	}
	defer webcam.Close()

	c.videoCapture = webcam
	c.isRunning = true

	if !webcam.IsOpened() {
		return fmt.Errorf("❌ не удалось открыть RTSP поток")
	}

	log.Println("✅ RTSP поток успешно открыт")

	// Получаем информацию о видео
	fps := webcam.Get(gocv.VideoCaptureFPS)
	width := webcam.Get(gocv.VideoCaptureFrameWidth)
	height := webcam.Get(gocv.VideoCaptureFrameHeight)

	log.Printf("📺 Параметры видео: %.0fx%.0f @ %.1f FPS", width, height, fps)

	// Основной цикл чтения кадров с таймаутом
	img := gocv.NewMat()
	defer img.Close()

	// Используем ticker для контроля частоты чтения
	ticker := time.NewTicker(time.Millisecond * 500) // читаем каждые 500мс
	defer ticker.Stop()

	frameCount := 0

	for {
		select {
		case <-ctx.Done():
			log.Println("🛑 Получен сигнал остановки камеры")
			c.isRunning = false
			return nil

		case <-ticker.C:
			// Читаем кадр с таймаутом
			readSuccess := make(chan bool, 1)
			
			go func() {
				ok := webcam.Read(&img)
				readSuccess <- ok
			}()

			// Ждем чтение кадра с таймаутом
			select {
			case ok := <-readSuccess:
				if !ok {
					log.Println("⚠️ Не удалось прочитать кадр, пытаемся переподключиться...")
					time.Sleep(time.Second * 2)
					continue
				}

				if img.Empty() {
					log.Println("⚠️ Получен пустой кадр")
					continue
				}

				frameCount++
				c.frameNum++

				// Обрабатываем только каждый N-й кадр
				if frameCount%c.config.FrameRate == 0 {
					log.Printf("📸 Обработка кадра #%d размером: %dx%d", c.frameNum, img.Cols(), img.Rows())
					c.processFrame(img.Clone())
				}

			case <-time.After(time.Second * 5):
				log.Println("⏰ Таймаут чтения кадра (5 сек)")
				continue
			}
		}
	}
}

// processFrame обрабатывает кадр
func (c *SimpleCamera) processFrame(frame gocv.Mat) {
	defer frame.Close()

	if c.config.SaveFrames {
		// Сохраняем кадр как изображение
		timestamp := time.Now().Format("2006-01-02_15-04-05")
		filename := filepath.Join(c.config.OutputDir, 
			fmt.Sprintf("frame_%s_%d.jpg", timestamp, c.frameNum))
		
		if ok := gocv.IMWrite(filename, frame); ok {
			log.Printf("💾 Кадр сохранен: %s", filename)
		} else {
			log.Printf("❌ Ошибка сохранения кадра: %s", filename)
		}
	}

	// Здесь будем добавлять детекцию объектов
	c.detectObjects(frame)
}

// detectObjects заготовка для детекции объектов
func (c *SimpleCamera) detectObjects(frame gocv.Mat) {
	// TODO: Здесь будет детекция людей и животных
	// Пока просто проверяем размер кадра
	if frame.Cols() > 0 && frame.Rows() > 0 {
		// Кадр готов для анализа
		log.Printf("🔍 Анализ кадра %dx%d (детекция в разработке)", frame.Cols(), frame.Rows())
	}
}

// checkFFmpeg проверяет доступность FFmpeg
func (c *SimpleCamera) checkFFmpeg() error {
	cmd := exec.Command("ffmpeg", "-version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("FFmpeg не найден. Установите: brew install ffmpeg")
	}
	return nil
}

// buildRTSPURL строит RTSP URL с аутентификацией
func (c *SimpleCamera) buildRTSPURL() string {
	if len(c.config.RTSPUrl) < 7 {
		return c.config.RTSPUrl // если URL короче "rtsp://"
	}
	return fmt.Sprintf("rtsp://%s:%s@%s",
		c.config.Username,
		c.config.Password,
		c.config.RTSPUrl[7:]) // убираем "rtsp://"
}

// maskURL маскирует пароль в URL для логирования
func (c *SimpleCamera) maskURL(url string) string {
	if len(url) > 30 {
		return url[:15] + "***" + url[len(url)-15:]
	}
	return "***"
}

// Stop останавливает камеру
func (c *SimpleCamera) Stop() {
	log.Println("🛑 Остановка камеры...")
	c.isRunning = false
	if c.videoCapture != nil {
		c.videoCapture.Close()
	}
}
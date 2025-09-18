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

// SimpleCamera использует FFmpeg для получения кадров через RTSP
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
	log.Printf("Подключение к камере через OpenCV: %s", c.maskURL(c.config.RTSPUrl))

	// Проверяем доступность FFmpeg
	if err := c.checkFFmpeg(); err != nil {
		return fmt.Errorf("FFmpeg недоступен: %v", err)
	}

	// Строим RTSP URL с аутентификацией
	rtspURL := c.buildRTSPURL()

	// Создаем VideoCapture
	webcam, err := gocv.OpenVideoCapture(rtspURL)
	if err != nil {
		return fmt.Errorf("ошибка открытия видеопотока: %v", err)
	}
	defer webcam.Close()

	c.videoCapture = webcam
	c.isRunning = true

	if !webcam.IsOpened() {
		return fmt.Errorf("не удалось открыть RTSP поток")
	}

	log.Println("RTSP поток успешно открыт")

	// Получаем информацию о видео
	fps := webcam.Get(gocv.VideoCaptureFPS)
	width := webcam.Get(gocv.VideoCaptureFrameWidth)
	height := webcam.Get(gocv.VideoCaptureFrameHeight)

	log.Printf("Параметры видео: %.0fx%.0f @ %.1f FPS", width, height, fps)

	// Основной цикл чтения кадров
	img := gocv.NewMat()
	defer img.Close()

	ticker := time.NewTicker(time.Millisecond * 200) // читаем кадры каждые 200мс
	defer ticker.Stop()

	frameCount := 0

	for {
		select {
		case <-ctx.Done():
			log.Println("Остановка камеры по сигналу...")
			c.isRunning = false
			return nil

		case <-ticker.C:
			if ok := webcam.Read(&img); !ok {
				log.Println("Не удалось прочитать кадр, пытаемся переподключиться...")
				time.Sleep(time.Second * 2)
				continue
			}

			if img.Empty() {
				continue
			}

			frameCount++
			c.frameNum++

			// Обрабатываем только каждый N-й кадр
			if frameCount%c.config.FrameRate == 0 {
				c.processFrame(img.Clone())
			}
		}
	}
}

// checkFFmpeg проверяет доступность FFmpeg
func (c *SimpleCamera) checkFFmpeg() error {
	cmd := exec.Command("ffmpeg", "-version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("FFmpeg не найден. Установите: sudo apt install ffmpeg")
	}
	return nil
}

// buildRTSPURL строит RTSP URL с аутентификацией
func (c *SimpleCamera) buildRTSPURL() string {
	if c.config.Username != "" && c.config.Password != "" {
		// Заменяем rtsp:// на rtsp://user:pass@
		if len(c.config.RTSPUrl) > 7 {
			return fmt.Sprintf("rtsp://%s:%s@%s", 
				c.config.Username, 
				c.config.Password, 
				c.config.RTSPUrl[7:]) // убираем "rtsp://"
		}
	}
	return c.config.RTSPUrl
}

// maskURL маскирует пароль в URL для логирования
func (c *SimpleCamera) maskURL(url string) string {
	if len(url) > 20 {
		return url[:10] + "***" + url[len(url)-10:]
	}
	return "***"
}

// processFrame обрабатывает кадр
func (c *SimpleCamera) processFrame(frame gocv.Mat) {
	defer frame.Close()

	log.Printf("Обработан кадр #%d размером: %dx%d", c.frameNum, frame.Cols(), frame.Rows())

	if c.config.SaveFrames {
		// Сохраняем кадр как изображение
		timestamp := time.Now().Format("2006-01-02_15-04-05")
		filename := filepath.Join(c.config.OutputDir, 
			fmt.Sprintf("frame_%s_%d.jpg", timestamp, c.frameNum))
		
		if ok := gocv.IMWrite(filename, frame); ok {
			log.Printf("Кадр сохранен: %s", filename)
		} else {
			log.Printf("Ошибка сохранения кадра: %s", filename)
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
	}
}

// Stop останавливает камеру
func (c *SimpleCamera) Stop() {
	log.Println("Остановка простой камеры...")
	c.isRunning = false
	if c.videoCapture != nil {
		c.videoCapture.Close()
	}
}
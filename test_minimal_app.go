package main

import (
	"fmt"
    "context"
    "log"
    "time"
    "gocv.io/x/gocv"
    "camera-detection-project/internal/config"
)

func main() {
    log.Println("=== 🎬 МИНИМАЛЬНОЕ ПРИЛОЖЕНИЕ ===")
    
    // Таймаут для безопасности
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    // Загружаем конфигурацию
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("❌ Ошибка конфигурации: %v", err)
    }
    
    // Строим RTSP URL
    rtspURL := buildRTSPURL(cfg)
    log.Printf("🔗 RTSP URL: %s", maskURL(rtspURL))
    
    // Подключаемся к камере
    log.Println("📹 Открываем видеопоток...")
    webcam, err := gocv.OpenVideoCapture(rtspURL)
    if err != nil {
        log.Fatalf("❌ Ошибка открытия потока: %v", err)
    }
    defer webcam.Close()
    
    if !webcam.IsOpened() {
        log.Fatalf("❌ Поток не открылся")
    }
    
    log.Println("✅ Поток открыт успешно!")
    
    // Читаем несколько кадров
    img := gocv.NewMat()
    defer img.Close()
    
    for i := 0; i < 10; i++ {
        select {
        case <-ctx.Done():
            log.Println("⏱️ Таймаут!")
            return
        default:
            log.Printf("📸 Читаем кадр #%d...", i+1)
            
            ok := webcam.Read(&img)
            if !ok {
                log.Printf("❌ Не удалось прочитать кадр #%d", i+1)
                continue
            }
            
            if img.Empty() {
                log.Printf("⚠️ Кадр #%d пустой", i+1)
                continue
            }
            
            log.Printf("✅ Кадр #%d: %dx%d", i+1, img.Cols(), img.Rows())
            
            // Пауза между кадрами
            time.Sleep(500 * time.Millisecond)
        }
    }
    
    log.Println("🎉 Тест минимального приложения успешно завершен!")
}

func buildRTSPURL(cfg *config.Config) string {
    if cfg.Camera.Username != "" && cfg.Camera.Password != "" {
        return fmt.Sprintf("rtsp://%s:%s@%s", 
            cfg.Camera.Username, 
            cfg.Camera.Password, 
            cfg.Camera.RTSPUrl[7:]) // убираем "rtsp://"
    }
    return cfg.Camera.RTSPUrl
}

func maskURL(url string) string {
    if len(url) > 30 {
        return url[:15] + "***" + url[len(url)-10:]
    }
    return "***"
}

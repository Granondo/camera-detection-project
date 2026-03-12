#!/usr/bin/env python3
"""
YOLOv8 Detection Service
Микросервис для детекции объектов с использованием YOLOv8
"""

import os
import logging
from flask import Flask, request, jsonify
from ultralytics import YOLO
import numpy as np
import time

# Настройка логирования
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

app = Flask(__name__)

class YOLODetector:
    def __init__(self, model_path="yolov8n.pt", confidence_threshold=0.5):
        """Инициализация детектора YOLO"""
        self.confidence_threshold = confidence_threshold
        logger.info(f"🔄 Loading YOLO model: {model_path}")
        
        try:
            self.model = YOLO(model_path)
            logger.info("✅ YOLO model loaded successfully")
            
            # Тестовый запуск для "разогрева" модели
            test_image = np.zeros((640, 640, 3), dtype=np.uint8)
            _ = self.model(test_image, verbose=False)
            logger.info("🚀 Model warmed up and ready")
            
        except Exception as e:
            logger.error(f"❌ Failed to load YOLO model: {e}")
            raise
    
    def detect(self, image_path):
        """Выполнить детекцию объектов на изображении"""
        logger.info(f"🔍 Starting detection for: {image_path}")
        
        if not os.path.exists(image_path):
            logger.error(f"❌ File not found: {image_path}")
            raise FileNotFoundError(f"Image not found: {image_path}")
        
        logger.info(f"✅ File exists, size: {os.path.getsize(image_path)} bytes")
        
        start_time = time.time()
        
        try:
            results = self.model(image_path, verbose=False)
            detection_time = time.time() - start_time
            logger.info(f"✅ YOLO completed in {detection_time:.2f}s")
            
            detections = []
            total_objects = 0
            
            for result in results:
                if result.boxes is not None and len(result.boxes) > 0:
                    logger.info(f"📦 Found {len(result.boxes)} boxes")
                    for box, conf, cls in zip(
                        result.boxes.xyxy.cpu().numpy(),
                        result.boxes.conf.cpu().numpy(), 
                        result.boxes.cls.cpu().numpy()
                    ):
                        if conf >= self.confidence_threshold:
                            detections.append({
                                'class': result.names[int(cls)],
                                'class_id': int(cls),
                                'confidence': float(conf),
                                'bbox': {
                                    'x1': float(box[0]),
                                    'y1': float(box[1]), 
                                    'x2': float(box[2]),
                                    'y2': float(box[3])
                                }
                            })
                            total_objects += 1
                else:
                    logger.info("📦 No boxes found in results")
            
            processing_time = time.time() - start_time
            
            return {
                'success': True,
                'image_path': image_path,
                'detections': detections,
                'total_objects': total_objects,
                'processing_time_ms': round(processing_time * 1000, 2),
                'model_confidence_threshold': self.confidence_threshold
            }
            
        except Exception as e:
            logger.error(f"❌ Detection failed for {image_path}: {e}")
            return {
                'success': False,
                'error': str(e),
                'image_path': image_path
            }

# Глобальный детектор - будет инициализирован при первом запросе
detector = None

def get_detector():
    """Ленивая инициализация детектора"""
    global detector
    if detector is None:
        model_path = os.getenv('YOLO_MODEL_PATH', 'yolov8n.pt')
        confidence = float(os.getenv('DETECTION_CONFIDENCE_THRESHOLD', '0.5'))
        logger.info(f"🚀 Initializing detector with model: {model_path}")
        detector = YOLODetector(model_path=model_path, confidence_threshold=confidence)
    return detector

@app.route('/health', methods=['GET'])
def health_check():
    """Проверка здоровья сервиса"""
    return jsonify({
        'status': 'healthy',
        'service': 'yolo-detection-service',
        'model_loaded': detector is not None
    })

@app.route('/detect', methods=['POST'])
def detect():
    """Основной эндпоинт для детекции объектов"""
    try:
        # Инициализация детектора при первом запросе
        det = get_detector()
        
        data = request.get_json()
        
        if not data or 'image_path' not in data:
            return jsonify({
                'success': False,
                'error': 'Missing image_path in request'
            }), 400
        
        image_path = data['image_path']
        logger.info(f"🔍 Processing detection request for: {os.path.basename(image_path)}")
        
        result = det.detect(image_path)
        
        if result['success']:
            logger.info(f"✅ Detection completed: {result['total_objects']} objects found in {result['processing_time_ms']}ms")
        
        return jsonify(result)
        
    except Exception as e:
        logger.error(f"❌ Request processing failed: {e}")
        return jsonify({
            'success': False,
            'error': f'Request processing failed: {str(e)}'
        }), 500

@app.route('/model/info', methods=['GET'])
def model_info():
    """Информация о загруженной модели"""
    det = get_detector()
    
    return jsonify({
        'model_loaded': True,
        'confidence_threshold': det.confidence_threshold,
        'available_classes': list(det.model.names.values()) if det.model else []
    })

if __name__ == '__main__':
    # Для локального запуска через python main.py
    host = os.getenv('HOST', '0.0.0.0')
    port = int(os.getenv('PORT', 5000))
    logger.info(f"🌟 Starting YOLO Detection Service (DEV MODE) on {host}:{port}")
    app.run(host=host, port=port, debug=True)
FROM python:3.11-slim

# Установка необходимых системных библиотек для OpenCV/OpenGL
RUN apt-get update && apt-get install -y \
    libglib2.0-0 \
    libgomp1 \
    libsm6 \
    libxext6 \
    libxrender-dev \
    libgl1-mesa-dev \
    libgles2-mesa-dev \
    libegl1-mesa-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Копируем requirements.txt
COPY requirements.txt .

# Устанавливаем PyTorch CPU-only
RUN pip install --no-cache-dir torch==2.0.1 torchvision==0.15.2 --index-url https://download.pytorch.org/whl/cpu

# Устанавливаем зависимости включая OpenCV headless
RUN pip install --no-cache-dir \
    ultralytics==8.0.196 \
    flask==3.0.0 \
    pillow==10.0.1 \
    numpy==1.24.3 \
    requests==2.31.0 \
    opencv-python-headless==4.8.1.78

# Копируем код
COPY . .

# Переменные окружения
ENV PYTHONUNBUFFERED=1
ENV YOLO_MODEL_PATH=yolov8n.pt
ENV CONFIDENCE_THRESHOLD=0.5

# Переменные для OpenCV без GUI
ENV QT_QPA_PLATFORM=offscreen
ENV DISPLAY=:99

EXPOSE 5000

CMD ["python", "main.py"]
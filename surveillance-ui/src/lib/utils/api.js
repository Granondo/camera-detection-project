// API Base URL
const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080';

// API Client с правильной обработкой 304
class ApiClient {
  constructor(baseUrl) {
    this.baseUrl = baseUrl;
  }

  async request(endpoint, options = {}) {
    const url = `${this.baseUrl}${endpoint}`;
    
    try {
      const response = await fetch(url, {
        ...options,
        headers: {
          'Content-Type': 'application/json',
          ...options.headers,
        },
        // ✅ Отключаем кэш для API запросов
        cache: 'no-cache',
      });

      console.log(`🔵 ${endpoint}:`, response.status); // Debug

      // ✅ Обработка 304 Not Modified
      if (response.status === 304) {
        console.log('⚠️  304 Not Modified - using cached data');
        // Возвращаем null или пустой объект
        return { success: false, cached: true };
      }

      if (!response.ok) {
        const text = await response.text();
        console.error('❌ API Error:', response.status, text);
        throw new Error(`HTTP ${response.status}: ${text.substring(0, 100)}`);
      }

      // Проверяем что есть тело ответа
      const contentLength = response.headers.get('content-length');
      if (contentLength === '0') {
        console.log('⚠️  Empty response body');
        return { success: false, empty: true };
      }

      // Проверяем Content-Type
      const contentType = response.headers.get('content-type');
      if (!contentType || !contentType.includes('application/json')) {
        const text = await response.text();
        console.error('❌ Non-JSON response:', text.substring(0, 200));
        throw new Error('Server returned non-JSON response');
      }

      const data = await response.json();
      return data;
      
    } catch (error) {
      console.error('💥 API Error:', error);
      throw error;
    }
  }

  // System
  async getHealth() {
    return this.request('/api/health');
  }

  async getStatus() {
    return this.request('/api/status');
  }

  async getStats() {
    return this.request('/api/stats');
  }

  async getCameraInfo() {
    return this.request('/api/camera');
  }

  // Recordings
  async getRecordings(limit = 20) {
    return this.request(`/api/recordings?limit=${limit}`);
  }

  async getRecording(id) {
    return this.request(`/api/recordings/${id}`);
  }

  getVideoUrl(id) {
    return `${this.baseUrl}/api/video/${id}`;
  }

  getDownloadUrl(id) {
    return `${this.baseUrl}/api/download/${id}`;
  }

  // Frames
  async getFrames(limit = 50, hasDetection = false) {
    const url = hasDetection 
      ? `/api/frames?has_detection=true&limit=${limit}`
      : `/api/frames?limit=${limit}`;
    return this.request(url);
  }

  async getFrame(id) {
    return this.request(`/api/frames/${id}`);
  }

  getImageUrl(id) {
    return `${this.baseUrl}/api/image/${id}`;
  }

  // Events
  async getEvents(limit = 20) {
    return this.request(`/api/events?limit=${limit}`);
  }
}

export const api = new ApiClient(API_BASE);
export default api;
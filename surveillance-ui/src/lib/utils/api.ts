import type { paths, components } from '../types/api';

// ✅ Типы из сгенерированного файла
type APIResponse = components['schemas']['api.APIResponse'];
type StatusResponse = components['schemas']['api.StatusResponse'];
type RecordingResponse = components['schemas']['api.RecordingResponse'];
type FrameResponse = components['schemas']['api.FrameResponse'];
type SystemStats = components['schemas']['api.SystemStats'];
type Camera = components['schemas']['storage.Camera'];
type Event = components['schemas']['storage.Event'];
type AnalyticsSummary = components['schemas']['api.AnalyticsSummaryResponse'];
type DetectionsHourly = components['schemas']['api.DetectionsHourlyResponse'];
type TopObjects = components['schemas']['api.TopObjectsResponse'];

// ✅ Generic wrapper для API ответов
interface TypedAPIResponse<T = unknown> extends APIResponse {
  data?: T;
}

// API Base URL
const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080';

// API Client с типизацией
class ApiClient {
  private baseUrl: string;

  constructor(baseUrl: string) {
    this.baseUrl = baseUrl;
  }

  private async request<T = unknown>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<TypedAPIResponse<T>> {
    const url = `${this.baseUrl}${endpoint}`;

    try {
      const response = await fetch(url, {
        ...options,
        headers: {
          'Content-Type': 'application/json',
          ...options.headers,
        },
        cache: 'no-cache',
      });

      console.log(`🔵 ${endpoint}:`, response.status);

      // ✅ Обработка 304 Not Modified
      if (response.status === 304) {
        console.log('⚠️  304 Not Modified - using cached data');
        return { success: false, cached: true } as TypedAPIResponse<T>;
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
        return { success: false } as TypedAPIResponse<T>;
      }

      // Проверяем Content-Type
      const contentType = response.headers.get('content-type');
      if (!contentType || !contentType.includes('application/json')) {
        const text = await response.text();
        console.error('❌ Non-JSON response:', text.substring(0, 200));
        throw new Error('Server returned non-JSON response');
      }

      const data = await response.json();
      return data as TypedAPIResponse<T>;
    } catch (error) {
      console.error('💥 API Error:', error);
      throw error;
    }
  }

  // ✅ System endpoints with types
  async getHealth(): Promise<TypedAPIResponse<void>> {
    return this.request('/api/health');
  }

  async getStatus(): Promise<TypedAPIResponse<StatusResponse>> {
    return this.request<StatusResponse>('/api/status');
  }

  async getStats(): Promise<TypedAPIResponse<{ [key: string]: number }>> {
    return this.request<{ [key: string]: number }>('/api/stats');
  }

  async getCameraInfo(): Promise<TypedAPIResponse<Camera>> {
    return this.request<Camera>('/api/camera');
  }

  // ✅ Recordings endpoints with types
  async getRecordings(limit = 20): Promise<TypedAPIResponse<RecordingResponse[]>> {
    return this.request<RecordingResponse[]>(`/api/recordings?limit=${limit}`);
  }

  async getRecording(id: number): Promise<TypedAPIResponse<RecordingResponse>> {
    return this.request<RecordingResponse>(`/api/recordings/${id}`);
  }

  getVideoUrl(id: number): string {
    return `${this.baseUrl}/api/video/${id}`;
  }

  getDownloadUrl(id: number): string {
    return `${this.baseUrl}/api/download/${id}`;
  }

  // ✅ Frames endpoints with types
  async getFrames(
    limit = 50,
    hasDetection = false
  ): Promise<TypedAPIResponse<FrameResponse[]>> {
    const url = hasDetection
      ? `/api/frames?has_detection=true&limit=${limit}`
      : `/api/frames?limit=${limit}`;
    return this.request<FrameResponse[]>(url);
  }

  async getFrame(id: number): Promise<TypedAPIResponse<FrameResponse>> {
    return this.request<FrameResponse>(`/api/frames/${id}`);
  }

  getImageUrl(id: number): string {
    return `${this.baseUrl}/api/image/${id}`;
  }

  // ✅ Events endpoints with types
  async getEvents(limit = 20): Promise<TypedAPIResponse<Event[]>> {
    return this.request<Event[]>(`/api/events?limit=${limit}`);
  }

  // ✅ Analytics endpoints with types
  async getAnalyticsSummary(): Promise<TypedAPIResponse<AnalyticsSummary>> {
    return this.request<AnalyticsSummary>('/api/analytics/summary');
  }

  async getDetectionsHourly(hours = 24): Promise<TypedAPIResponse<DetectionsHourly[]>> {
    return this.request<DetectionsHourly[]>(`/api/analytics/detections/hourly?hours=${hours}`);
  }

  async getTopObjects(
    days = 7,
    limit = 10
  ): Promise<TypedAPIResponse<TopObjects[]>> {
    return this.request<TopObjects[]>(
      `/api/analytics/top-objects?days=${days}&limit=${limit}`
    );
  }

  async getDetectionsCount(): Promise<TypedAPIResponse<{ [key: string]: number }>> {
    return this.request<{ [key: string]: number }>('/api/analytics/detections/count');
  }

  // ✅ Cache endpoint
  async clearCache(): Promise<TypedAPIResponse<void>> {
    return this.request('/api/cache/clear', { method: 'POST' });
  }
}

export const api = new ApiClient(API_BASE);
export default api;

// ✅ Export types for use in components
export type {
  APIResponse,
  StatusResponse,
  RecordingResponse,
  FrameResponse,
  SystemStats,
  Camera,
  Event,
  AnalyticsSummary,
  DetectionsHourly,
  TopObjects,
  TypedAPIResponse,
};
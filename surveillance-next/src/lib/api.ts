const getBaseUrl = () => {
  if (typeof window !== 'undefined') {
    return ''
  }
  return process.env.API_URL || 'http://api-server:8080'
}

const API_BASE = getBaseUrl()

export interface Event {
  id: number
  camera_id: number
  event_type: string
  severity: string
  title: string
  message: string
  metadata: unknown | null
  notified: boolean
  resolved: boolean
  timestamp: string
  created_at: string
  resolved_at: string | null
}

export interface Frame {
  id: number
  camera_id: number
  timestamp: string
  image_path: string
  detections: Detection[]
}

export interface Detection {
  id?: number
  type: string
  confidence: number
  bbox: {
    x: number
    y: number
    width: number
    height: number
  }
}

interface ApiResponse<T> {
  success: boolean
  data: T
}

export async function getEvents(params?: {
  page?: number
  perPage?: number
  type?: string
  cameraId?: number
}): Promise<{ events: Event[]; total: number }> {
  const searchParams = new URLSearchParams()
  if (params?.page) searchParams.set('page', params.page.toString())
  if (params?.perPage) searchParams.set('per_page', params.perPage.toString())
  if (params?.type) searchParams.set('type', params.type)
  if (params?.cameraId) searchParams.set('camera_id', params.cameraId.toString())

  const url = `${API_BASE}/api/events?${searchParams}`
  const res = await fetch(url, {
    next: { revalidate: 30 },
  })

  if (!res.ok) {
    return { events: [], total: 0 }
  }

  const json: ApiResponse<Event[]> = await res.json()
  return {
    events: json.data || [],
    total: json.data?.length || 0,
  }
}

export async function getEvent(id: string): Promise<Event | null> {
  const res = await fetch(`${API_BASE}/api/events/${id}`, {
    next: { revalidate: 30 },
  })

  if (!res.ok) {
    return null
  }

  const json: ApiResponse<Event> = await res.json()
  return json.data || null
}

export async function getFrames(params?: {
  page?: number
  perPage?: number
  cameraId?: number
  minConfidence?: number
}): Promise<{ frames: Frame[]; total: number }> {
  const searchParams = new URLSearchParams()
  if (params?.page) searchParams.set('page', params.page.toString())
  if (params?.perPage) searchParams.set('per_page', params.perPage.toString())
  if (params?.cameraId) searchParams.set('camera_id', params.cameraId.toString())
  if (params?.minConfidence) searchParams.set('min_confidence', params.minConfidence.toString())

  const url = `${API_BASE}/api/frames?${searchParams}`
  const res = await fetch(url, {
    next: { revalidate: 30 },
  })

  if (!res.ok) {
    return { frames: [], total: 0 }
  }

  const json: ApiResponse<Frame[]> = await res.json()
  return {
    frames: json.data || [],
    total: json.data?.length || 0,
  }
}

export async function getFrame(id: string): Promise<Frame | null> {
  const res = await fetch(`${API_BASE}/api/frames/${id}`, {
    next: { revalidate: 30 },
  })

  if (!res.ok) {
    return null
  }

  const json: ApiResponse<Frame> = await res.json()
  return json.data || null
}

export function getImageUrl(path: string): string {
  if (path.startsWith('http')) return path
  return `${API_BASE}/images/${path}`
}
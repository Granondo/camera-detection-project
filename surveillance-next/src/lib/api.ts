const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'

export interface Detection {
  id: string
  type: string
  confidence: number
  bbox: {
    x: number
    y: number
    width: number
    height: number
  }
}

export interface Frame {
  id: string
  cameraId: string
  timestamp: string
  imagePath: string
  detections: Detection[]
}

export interface Event {
  id: string
  type: string
  startTime: string
  endTime: string | null
  cameraId: string
  frames: Frame[]
  status: 'active' | 'completed'
}

export interface EventsResponse {
  events: Event[]
  total: number
  page: number
  perPage: number
}

export interface FramesResponse {
  frames: Frame[]
  total: number
  page: number
  perPage: number
}

export async function getEvents(params?: {
  page?: number
  perPage?: number
  type?: string
  cameraId?: string
}): Promise<EventsResponse> {
  const searchParams = new URLSearchParams()
  if (params?.page) searchParams.set('page', params.page.toString())
  if (params?.perPage) searchParams.set('per_page', params.perPage.toString())
  if (params?.type) searchParams.set('type', params.type)
  if (params?.cameraId) searchParams.set('camera_id', params.cameraId)

  const res = await fetch(`${API_BASE}/api/events?${searchParams}`, {
    next: { revalidate: 30 },
  })

  if (!res.ok) throw new Error('Failed to fetch events')
  return res.json()
}

export async function getEvent(id: string): Promise<Event> {
  const res = await fetch(`${API_BASE}/api/events/${id}`, {
    next: { revalidate: 30 },
  })

  if (!res.ok) throw new Error('Failed to fetch event')
  return res.json()
}

export async function getFrames(params?: {
  page?: number
  perPage?: number
  cameraId?: string
  minConfidence?: number
}): Promise<FramesResponse> {
  const searchParams = new URLSearchParams()
  if (params?.page) searchParams.set('page', params.page.toString())
  if (params?.perPage) searchParams.set('per_page', params.perPage.toString())
  if (params?.cameraId) searchParams.set('camera_id', params.cameraId)
  if (params?.minConfidence) searchParams.set('min_confidence', params.minConfidence.toString())

  const res = await fetch(`${API_BASE}/api/frames?${searchParams}`, {
    next: { revalidate: 30 },
  })

  if (!res.ok) throw new Error('Failed to fetch frames')
  return res.json()
}

export async function getFrame(id: string): Promise<Frame> {
  const res = await fetch(`${API_BASE}/api/frames/${id}`, {
    next: { revalidate: 30 },
  })

  if (!res.ok) throw new Error('Failed to fetch frame')
  return res.json()
}

export function getImageUrl(path: string): string {
  return `${API_BASE}/images/${path}`
}
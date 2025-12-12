const getBaseUrl = () => {
  if (typeof window !== "undefined") {
    return "";
  }
  return process.env.API_URL || "http://api-server:8080";
};

const API_BASE = getBaseUrl();

export interface Event {
  id: number;
  camera_id: number;
  frame_id: number | null;
  event_type: string;
  severity: string;
  title: string;
  message: string;
  metadata: unknown | null;
  notified: boolean;
  resolved: boolean;
  timestamp: string;
  created_at: string;
  resolved_at: string | null;
}

export interface Frame {
  id: number;
  camera_id: number;
  timestamp: string;
  image_path: string;
  detections: Detection[];
}

export interface Detection {
  id?: number;
  type: string;
  confidence: number;
  bbox: {
    x: number;
    y: number;
    width: number;
    height: number;
  };
}

interface ApiResponse<T> {
  success: boolean;
  data: T;
}

export interface PaginatedEventsResponse {
  events: Event[];
  total: number;
  page: number;
  per_page: number;
  total_pages: number;
}

export async function getEvents(params?: {
  page?: number;
  perPage?: number;
  type?: string;
  cameraId?: number;
}): Promise<PaginatedEventsResponse> {
  const searchParams = new URLSearchParams();
  if (params?.page) searchParams.set("page", params.page.toString());
  if (params?.perPage) searchParams.set("per_page", params.perPage.toString());
  if (params?.type) searchParams.set("type", params.type);
  if (params?.cameraId)
    searchParams.set("camera_id", params.cameraId.toString());

  const url = `${API_BASE}/api/events?${searchParams}`;
  const res = await fetch(url, {
    next: { revalidate: 30 },
  });

  if (!res.ok) {
    return {
      events: [],
      total: 0,
      page: 1,
      per_page: 20,
      total_pages: 0,
    };
  }

  const json: ApiResponse<PaginatedEventsResponse> = await res.json();
  return (
    json.data || {
      events: [],
      total: 0,
      page: 1,
      per_page: 20,
      total_pages: 0,
    }
  );
}

export async function getEvent(id: string): Promise<Event | null> {
  const res = await fetch(`${API_BASE}/api/events/${id}`, {
    next: { revalidate: 30 },
  });

  if (!res.ok) {
    return null;
  }

  const json: ApiResponse<Event> = await res.json();
  return json.data || null;
}

export interface SearchParams {
  q?: string;
  severity?: string;
  type?: string;
  limit?: number;
}

export async function searchEvents(params?: SearchParams): Promise<Event[]> {
  const searchParams = new URLSearchParams();
  if (params?.q) searchParams.set("q", params.q);
  if (params?.severity) searchParams.set("severity", params.severity);
  if (params?.type) searchParams.set("type", params.type);
  if (params?.limit) searchParams.set("limit", params.limit.toString());

  const url = `${API_BASE}/api/search?${searchParams}`;
  const res = await fetch(url, {
    cache: "no-store",
  });

  if (!res.ok) {
    return [];
  }

  const json: ApiResponse<Event[]> = await res.json();
  return json.data || [];
}

export interface PaginatedFramesResponse {
  frames: Frame[];
  total: number;
  page: number;
  per_page: number;
  total_pages: number;
}

export async function getFrames(params?: {
  page?: number;
  perPage?: number;
  cameraId?: number;
  minConfidence?: number;
}): Promise<PaginatedFramesResponse> {
  const searchParams = new URLSearchParams();
  if (params?.page) searchParams.set("page", params.page.toString());
  if (params?.perPage) searchParams.set("per_page", params.perPage.toString());
  if (params?.cameraId)
    searchParams.set("camera_id", params.cameraId.toString());
  if (params?.minConfidence)
    searchParams.set("min_confidence", params.minConfidence.toString());

  const url = `${API_BASE}/api/frames?${searchParams}`;
  const res = await fetch(url, {
    next: { revalidate: 30 },
  });

  const emptyResponse = {
    frames: [],
    total: 0,
    page: 1,
    per_page: 16,
    total_pages: 0,
  };

  if (!res.ok) {
    return emptyResponse;
  }

  interface FrameWrapper {
    frame: Frame;
    file_exists: boolean;
    image_url: string;
  }
  
  // The API returns a different structure for paginated frames
  interface PaginatedFramesApiResponse {
    frames: FrameWrapper[];
    total: number;
    page: number;
    per_page: number;
    total_pages: number;
  }

  const json: ApiResponse<PaginatedFramesApiResponse> = await res.json();
  const data = json.data;

  if (!data) {
    return emptyResponse;
  }

  const frames = (data.frames || []).map((item) => ({
    ...item.frame,
    image_path: item.image_url,
  }));

  return {
    frames,
    total: data.total,
    page: data.page,
    per_page: data.per_page,
    total_pages: data.total_pages,
  };
}

export async function getFrame(id: string): Promise<Frame | null> {
  const res = await fetch(`${API_BASE}/api/frames/${id}`, {
    next: { revalidate: 30 },
  });

  if (!res.ok) {
    return null;
  }

  interface FrameWrapper {
    frame: Frame;
    file_exists: boolean;
    image_url: string;
  }

  const json: ApiResponse<FrameWrapper> = await res.json();
  if (!json.data) return null;

  return {
    ...json.data.frame,
    image_path: json.data.image_url,
  };
}

export function getImageUrl(path: string): string {
  // If already a full URL, return as-is
  if (path.startsWith("http")) return path;

  // If path is already an absolute API path (starts with /api/),
  // return it as-is for browser to use (relative URL)
  // This works for both SSR and client-side rendering
  if (path.startsWith("/api/")) {
    return path;
  }

  // For client-side, use the public API URL
  // For server-side, return relative path
  const baseUrl =
    typeof window !== "undefined"
      ? process.env.NEXT_PUBLIC_API_URL || ""
      : "";

  // If we have a baseUrl and path doesn't start with /, prepend baseUrl
  if (baseUrl && !path.startsWith("/")) {
    return `${baseUrl}/images/${path}`;
  }

  // Otherwise, return relative path
  return path.startsWith("/") ? path : `/images/${path}`;
}

export interface Link {
  id: number;
  code: string;
  short_url: string;
  original_url: string;
  clicks: number;
  expires_at?: string;
  created_at: string;
}

export interface ApiResponse<T> {
  data?: T;
  errors?: string;
}

export interface Link {
  id: number;
  code: string;
  short_url: string;
  original_url: string;
  clicks: number;
  expires_at?: string;
  created_at: string;
  tags?: string[];
}

export interface User {
  id: number;
  name: string;
  email: string;
}

export interface LinkPage {
  items: Link[];
  total: number;
  page: number;
  size: number;
}

export interface LinkListFilters {
  q?: string;
  status?: "" | "active" | "expired";
  tag?: string;
}

export interface AnalyticsDailyClicks {
  date: string;
  clicks: number;
}

export interface AnalyticsBreakdown {
  name: string;
  count: number;
}

export interface LinkAnalytics {
  days: number;
  since: string;
  total_clicks: number;
  unique_visitors: number;
  location_available: boolean;
  clicks_by_day: AnalyticsDailyClicks[];
  referrers: AnalyticsBreakdown[];
  countries: AnalyticsBreakdown[];
  devices: AnalyticsBreakdown[];
}

export interface ApiResponse<T> {
  data?: T;
  errors?: string;
}

export interface FunnelStage {
  event_type: string;
  count: number;
  unique_orders: number;
  rate: number;
}

export interface CancellationRow {
  city: string;
  count: number;
  day: string;
}

export interface VolumePoint {
  day: string;
  count: number;
}

export interface RecentEvent {
  event_id: string;
  event_type: string;
  city: string;
  timestamp: string;
  order_id: string;
  buyer_hash: string;
  amount: number;
}

export interface HealthStatus {
  status: string;
  service: string;
  clickhouse: string;
  timestamp: string;
}

export interface User {
  id: number;
  email: string;
  name: string;
  role: string;
}

export interface AuthResponse {
  token: string;
  user: User;
}

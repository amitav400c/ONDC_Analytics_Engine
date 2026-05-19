import { useQuery } from '@tanstack/react-query';
import type { FunnelStage, CancellationRow, VolumePoint, RecentEvent, HealthStatus, AuthResponse } from '../types';

const API_BASE = '/api/v1';

function getToken(): string | null {
  return localStorage.getItem('ondc_token');
}

async function apiFetch<T>(path: string): Promise<T> {
  const token = getToken();
  const res = await fetch(`${API_BASE}${path}`, {
    headers: token ? { Authorization: `Bearer ${token}` } : {},
  });
  if (res.status === 401) {
    localStorage.removeItem('ondc_token');
    localStorage.removeItem('ondc_user');
    window.location.reload();
    throw new Error('Unauthorized');
  }
  if (!res.ok) throw new Error(`API error: ${res.status}`);
  return res.json();
}

export async function login(email: string, password: string): Promise<AuthResponse> {
  const res = await fetch(`${API_BASE}/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password }),
  });
  if (!res.ok) throw new Error('Invalid credentials');
  return res.json();
}

export function useHealth() {
  return useQuery<HealthStatus>({
    queryKey: ['health'],
    queryFn: () => fetch(`${API_BASE}/health`).then(r => r.json()),
    refetchInterval: 10000,
  });
}

export function useFunnel() {
  return useQuery<{ stages: FunnelStage[] }>({
    queryKey: ['funnel'],
    queryFn: () => apiFetch('/metrics/funnel'),
    refetchInterval: 15000,
  });
}

export function useCancellations(city?: string) {
  return useQuery<{ cancellations: CancellationRow[] }>({
    queryKey: ['cancellations', city],
    queryFn: () => apiFetch(`/metrics/cancellations${city ? `?city=${city}` : ''}`),
    refetchInterval: 15000,
  });
}

export function useVolume(days = 7) {
  return useQuery<{ volume: VolumePoint[] }>({
    queryKey: ['volume', days],
    queryFn: () => apiFetch(`/metrics/volume?days=${days}`),
    refetchInterval: 15000,
  });
}

export function useRecentEvents(limit = 20) {
  return useQuery<{ events: RecentEvent[] }>({
    queryKey: ['recent', limit],
    queryFn: () => apiFetch(`/events/recent?limit=${limit}`),
    refetchInterval: 5000,
  });
}

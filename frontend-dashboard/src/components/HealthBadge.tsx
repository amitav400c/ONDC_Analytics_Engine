import { useHealth } from '../api/client';

export default function HealthBadge() {
  const { data, isError } = useHealth();
  const isUp = data?.status === 'ok' && data?.clickhouse === 'up';

  return (
    <div className="flex items-center gap-2 text-sm">
      <span
        className={`w-2 h-2 rounded-full ${
          isError ? 'bg-red-500' : isUp ? 'bg-emerald-400 animate-pulse-soft' : 'bg-yellow-400'
        }`}
      />
      <span className="text-white/40">
        {isError ? 'Offline' : isUp ? 'All Systems Up' : 'Degraded'}
      </span>
    </div>
  );
}

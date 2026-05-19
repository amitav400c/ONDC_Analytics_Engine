
import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Legend
} from 'recharts';
import { useLatency } from '../api/client';

export function LatencyChart() {
  const { data: rawData, isLoading } = useLatency(60);

  const data = rawData?.latency?.map((p: any) => ({
    ...p,
    displayTime: new Date(p.time).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  })) || [];

  if (isLoading && data.length === 0) {
    return (
      <div className="bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-100 dark:border-gray-700 p-6 animate-pulse h-80 flex items-center justify-center">
        <span className="text-gray-400">Loading Latency Metrics...</span>
      </div>
    );
  }

  return (
    <div className="bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-100 dark:border-gray-700 p-6 transition-all hover:shadow-md">
      <h3 className="text-lg font-semibold text-gray-800 dark:text-gray-100 mb-6">
        System Processing Latency (last 60m)
      </h3>
      <div className="h-72 w-full">
        <ResponsiveContainer width="100%" height="100%">
          <AreaChart data={data} margin={{ top: 10, right: 30, left: 0, bottom: 0 }}>
            <defs>
              <linearGradient id="colorTotal" x1="0" y1="0" x2="0" y2="1">
                <stop offset="5%" stopColor="#ef4444" stopOpacity={0.3}/>
                <stop offset="95%" stopColor="#ef4444" stopOpacity={0}/>
              </linearGradient>
              <linearGradient id="colorSandbox" x1="0" y1="0" x2="0" y2="1">
                <stop offset="5%" stopColor="#3b82f6" stopOpacity={0.3}/>
                <stop offset="95%" stopColor="#3b82f6" stopOpacity={0}/>
              </linearGradient>
            </defs>
            <XAxis 
              dataKey="displayTime" 
              stroke="#9ca3af" 
              fontSize={12}
              tickLine={false}
              axisLine={false}
            />
            <YAxis 
              stroke="#9ca3af" 
              fontSize={12}
              tickLine={false}
              axisLine={false}
              tickFormatter={(value) => `${value}ms`}
            />
            <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#e5e7eb" className="dark:stroke-gray-700" />
            <Tooltip
              contentStyle={{
                backgroundColor: 'rgba(255, 255, 255, 0.95)',
                border: 'none',
                borderRadius: '8px',
                boxShadow: '0 4px 6px -1px rgba(0, 0, 0, 0.1)',
                color: '#1f2937'
              }}
              itemStyle={{ color: '#1f2937', fontWeight: 500 }}
              labelStyle={{ color: '#6b7280', marginBottom: '4px' }}
            />
            <Legend verticalAlign="top" height={36} wrapperStyle={{ fontSize: '12px' }}/>
            <Area 
              type="monotone" 
              dataKey="avg_total_ms" 
              name="Total Latency" 
              stroke="#ef4444" 
              strokeWidth={2}
              fillOpacity={1} 
              fill="url(#colorTotal)" 
            />
            <Area 
              type="monotone" 
              dataKey="avg_sandbox_ms" 
              name="Sandbox Latency" 
              stroke="#3b82f6" 
              strokeWidth={2}
              fillOpacity={1} 
              fill="url(#colorSandbox)" 
            />
            <Area 
              type="monotone" 
              dataKey="avg_kafka_ms" 
              name="Kafka Latency" 
              stroke="#10b981" 
              strokeWidth={2}
              fillOpacity={0} 
            />
          </AreaChart>
        </ResponsiveContainer>
      </div>
    </div>
  );
}

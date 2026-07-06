import { useCallback, useEffect, useRef, useState } from 'react';

import { connectToEvents } from '@/lib/api';

export interface WorkerLog {
  id: string;
  message: string;
  timestamp: number;
}

export interface WorkerState {
  id: string;
  status: string;
  message: string;
  lastUpdate: number;
  logs: WorkerLog[];
  workCount?: number;
}

export interface SSEState {
  progress: number;
  status: string;
  message: string;
  isConnected: boolean;
  error: string | null;
  workers: Record<string, WorkerState>;
  chartData: Array<Record<string, number | string>>;
}

const initialState: SSEState = {
  progress: 0,
  status: '',
  message: '',
  isConnected: false,
  error: null,
  workers: {},
  chartData: [],
};

export function useSSE(taskID: string | null) {
  const [state, setState] = useState<SSEState>(initialState);
  const [retryCount, setRetryCount] = useState(0);

  const processedTimestamps = useRef<Set<string>>(new Set());
  const statusRef = useRef(state.status);
  statusRef.current = state.status;

  const reset = useCallback(() => {
    setState({
      ...initialState,
      chartData: [],
    });
    setRetryCount(0);
    processedTimestamps.current.clear();
  }, []);

  useEffect(() => {
    if (!taskID) {
      setState(initialState);
      return;
    }

    if (
      statusRef.current === 'JOB_COMPLETED' ||
      statusRef.current === 'JOB_FAILED'
    ) {
      return;
    }

    let timer: NodeJS.Timeout;
    const eventSource = connectToEvents(taskID);

    eventSource.onopen = () => {
      setState((prev) => ({ ...prev, isConnected: true, error: null }));
      setRetryCount(0);
    };

    eventSource.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data) as {
          progress?: number;
          status?: string;
          message?: string;
          worker_id?: string;
          timestamp?: number;
        };

        const tsKey = data.timestamp
          ? String(data.timestamp)
          : `${Date.now()}-${Math.random()}`;

        if (data.timestamp && processedTimestamps.current.has(tsKey)) {
          return;
        }
        processedTimestamps.current.add(tsKey);

        setState((prev) => {
          const nextProgress = Math.max(prev.progress, data.progress ?? 0);

          const nextWorkers = { ...prev.workers };
          let nextChartData = [...prev.chartData];

          if (data.worker_id) {
            const currentWorker = prev.workers[data.worker_id] || {
              id: data.worker_id,
              status: 'unknown',
              message: '',
              lastUpdate: Date.now(),
              logs: [],
              workCount: 0,
            };

            const logTimestamp = data.timestamp ? data.timestamp : Date.now();
            const newLogs: WorkerLog[] = data.message
              ? [
                  ...currentWorker.logs,
                  { id: tsKey, message: data.message, timestamp: logTimestamp },
                ].slice(-50)
              : currentWorker.logs;

            const isWorkMessage =
              data.message &&
              (data.message.includes('fragmento') ||
                data.message.includes('frame') ||
                data.message.includes('Segmento') ||
                data.message.includes('Procesando'));

            const currentWorkerWorkCount = currentWorker.workCount || 0;
            const nextWorkCount = isWorkMessage
              ? currentWorkerWorkCount + 1
              : currentWorkerWorkCount;

            nextWorkers[data.worker_id] = {
              ...currentWorker,
              status: data.status ?? currentWorker.status,
              message: data.message ?? currentWorker.message,
              lastUpdate: Date.now(),
              logs: newLogs,
              workCount: nextWorkCount,
            };

            if (isWorkMessage) {
              let ms = logTimestamp;
              if (logTimestamp > 99999999999999) {
                ms = Math.floor(logTimestamp / 1000000);
              } else if (logTimestamp < 9999999999) {
                ms = logTimestamp * 1000;
              }
              const timestampStr = new Date(ms).toLocaleTimeString([], {
                hour: '2-digit',
                minute: '2-digit',
                second: '2-digit',
                hour12: false,
              });

              const currentCounts: Record<string, number> = {};
              for (const wId of Object.keys(nextWorkers)) {
                currentCounts[`NODE-${wId.slice(0, 4)}`] =
                  nextWorkers[wId].workCount || 0;
              }

              nextChartData.push({
                time: timestampStr,
                ...currentCounts,
              });

              if (nextChartData.length > 100) {
                nextChartData = nextChartData.slice(-100);
              }
            }
          }

          return {
            ...prev,
            progress: nextProgress,
            status: data.status ?? prev.status,
            message: data.message ?? prev.message,
            workers: nextWorkers,
            chartData: nextChartData,
          };
        });
      } catch {}
    };

    eventSource.onerror = () => {
      eventSource.close();
      setState((prev) => ({ ...prev, isConnected: false }));

      const delay = Math.min(1000 * 2 ** retryCount, 30000);
      timer = setTimeout(() => {
        setRetryCount((c) => c + 1);
      }, delay);
    };

    return () => {
      eventSource.close();
      if (timer) clearTimeout(timer);
    };
  }, [taskID, retryCount]);

  return { ...state, reset };
}

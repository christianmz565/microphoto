import { useCallback, useEffect, useRef, useState } from 'react';

import { connectToEvents } from '@/lib/api';

export interface WorkerLog {
  id: string;
  message: string;
}

export interface WorkerState {
  id: string;
  status: string;
  message: string;
  lastUpdate: number;
  logs: WorkerLog[];
}

export interface SSEState {
  progress: number;
  status: string;
  message: string;
  isConnected: boolean;
  error: string | null;
  workers: Record<string, WorkerState>;
}

const initialState: SSEState = {
  progress: 0,
  status: '',
  message: '',
  isConnected: false,
  error: null,
  workers: {},
};

export function useSSE(taskID: string | null) {
  const [state, setState] = useState<SSEState>(initialState);
  const [retryCount, setRetryCount] = useState(0);

  const processedTimestamps = useRef<Set<string>>(new Set());
  const statusRef = useRef(state.status);
  statusRef.current = state.status;

  const reset = useCallback(() => {
    setState(initialState);
    setRetryCount(0);
    processedTimestamps.current.clear();
  }, []);

  useEffect(() => {
    if (!taskID) return;

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
          if (data.worker_id) {
            const currentWorker = prev.workers[data.worker_id] || {
              id: data.worker_id,
              status: 'unknown',
              message: '',
              lastUpdate: Date.now(),
              logs: [],
            };

            const newLogs: WorkerLog[] = data.message
              ? [
                  ...currentWorker.logs,
                  { id: tsKey, message: data.message },
                ].slice(-50)
              : currentWorker.logs;

            nextWorkers[data.worker_id] = {
              ...currentWorker,
              status: data.status ?? currentWorker.status,
              message: data.message ?? currentWorker.message,
              lastUpdate: Date.now(),
              logs: newLogs,
            };
          }

          return {
            ...prev,
            progress: nextProgress,
            status: data.status ?? prev.status,
            message: data.message ?? prev.message,
            workers: nextWorkers,
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

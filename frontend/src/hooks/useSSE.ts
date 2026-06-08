import { useCallback, useEffect, useState } from 'react';

import { connectToEvents } from '@/lib/api';

export interface SSEState {
  progress: number;
  status: string;
  message: string;
  isConnected: boolean;
  error: string | null;
}

const initialState: SSEState = {
  progress: 0,
  status: '',
  message: '',
  isConnected: false,
  error: null,
};

export function useSSE(taskID: string | null) {
  const [state, setState] = useState<SSEState>(initialState);

  const reset = useCallback(() => {
    setState(initialState);
  }, []);

  useEffect(() => {
    if (!taskID) return;

    const eventSource = connectToEvents(taskID);

    eventSource.onopen = () => {
      setState((prev) => ({ ...prev, isConnected: true }));
    };

    eventSource.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data) as {
          progress?: number;
          status?: string;
          message?: string;
        };
        setState((prev) => ({
          ...prev,
          progress: data.progress ?? prev.progress,
          status: data.status ?? prev.status,
          message: data.message ?? prev.message,
        }));
      } catch {
        // ignore malformed events
      }
    };

    eventSource.onerror = () => {
      setState((prev) => ({
        ...prev,
        isConnected: false,
        error: 'Connection lost',
      }));
      eventSource.close();
    };

    return () => {
      eventSource.close();
    };
  }, [taskID]);

  return { ...state, reset };
}

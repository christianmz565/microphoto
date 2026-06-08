import { useCallback, useEffect, useState } from "react";

import { connectToEvents } from "@/lib/api";

export interface SSEState {
  progress: number;
  status: string;
  message: string;
  isConnected: boolean;
  error: string | null;
}

const initialState: SSEState = {
  progress: 0,
  status: "",
  message: "",
  isConnected: false,
  error: null,
};

export function useSSE(taskID: string | null) {
  const [state, setState] = useState<SSEState>(initialState);
  const [retryCount, setRetryCount] = useState(0);

  const reset = useCallback(() => {
    setState(initialState);
    setRetryCount(0);
  }, []);

  useEffect(() => {
    if (!taskID) return;

    if (state.status === "JOB_COMPLETED" || state.status === "JOB_FAILED") {
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
        };
        setState((prev) => ({
          ...prev,
          progress: data.progress ?? prev.progress,
          status: data.status ?? prev.status,
          message: data.message ?? prev.message,
        }));
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
  }, [taskID, retryCount, state.status]);

  return { ...state, reset };
}

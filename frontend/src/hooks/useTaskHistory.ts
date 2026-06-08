import { useCallback, useEffect, useState } from 'react';

const STORAGE_KEY = 'microphoto:history';
const MAX_ENTRIES = 20;

export interface TaskEntry {
  taskID: string;
  filename: string;
  filterType: string;
  timestamp: number;
  status: 'processing' | 'completed' | 'failed';
}

function loadHistory(): TaskEntry[] {
  if (typeof window === 'undefined') return [];
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    return raw ? (JSON.parse(raw) as TaskEntry[]) : [];
  } catch {
    return [];
  }
}

function saveHistory(tasks: TaskEntry[]) {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(tasks));
}

export function useTaskHistory() {
  const [tasks, setTasks] = useState<TaskEntry[]>([]);

  useEffect(() => {
    setTasks(loadHistory());
  }, []);

  const addTask = useCallback((entry: Omit<TaskEntry, 'timestamp'>) => {
    setTasks((prev) => {
      const updated = [{ ...entry, timestamp: Date.now() }, ...prev].slice(
        0,
        MAX_ENTRIES,
      );
      saveHistory(updated);
      return updated;
    });
  }, []);

  const updateStatus = useCallback(
    (taskID: string, status: TaskEntry['status']) => {
      setTasks((prev) => {
        const updated = prev.map((t) =>
          t.taskID === taskID ? { ...t, status } : t,
        );
        saveHistory(updated);
        return updated;
      });
    },
    [],
  );

  const removeTask = useCallback((taskID: string) => {
    setTasks((prev) => {
      const updated = prev.filter((t) => t.taskID !== taskID);
      saveHistory(updated);
      return updated;
    });
  }, []);

  const clearHistory = useCallback(() => {
    setTasks([]);
    localStorage.removeItem(STORAGE_KEY);
  }, []);

  return { tasks, addTask, updateStatus, removeTask, clearHistory };
}

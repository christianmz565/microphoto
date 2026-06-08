import { IconClock, IconDownload, IconPhoto, IconTrash } from '@tabler/icons-react';
import { useState } from 'react';

import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Separator } from '@/components/ui/separator';
import { useTaskHistory } from '@/hooks/useTaskHistory';
import { getResult } from '@/lib/api';

export function TaskHistory() {
  const { tasks, removeTask, clearHistory } = useTaskHistory();
  const [downloading, setDownloading] = useState<string | null>(null);

  const handleDownload = async (taskID: string) => {
    setDownloading(taskID);
    try {
      const blob = await getResult(taskID);
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `processed-${taskID.slice(0, 8)}.png`;
      a.click();
      URL.revokeObjectURL(url);
    } catch {
      // ignore download errors
    } finally {
      setDownloading(null);
    }
  };

  if (tasks.length === 0) {
    return (
      <Card className="mx-auto w-full max-w-md">
        <CardContent className="flex flex-col items-center gap-3 py-12 text-muted-foreground">
          <IconPhoto className="size-10" />
          <p>No tasks yet</p>
          <p className="text-xs">Upload an image to get started</p>
        </CardContent>
      </Card>
    );
  }

  return (
    <div className="mx-auto flex w-full max-w-md flex-col gap-3">
      <div className="flex items-center justify-between">
        <span className="text-sm text-muted-foreground">
          {tasks.length} task{tasks.length !== 1 ? 's' : ''}
        </span>
        <Button variant="ghost" size="sm" onClick={clearHistory}>
          <IconTrash className="size-3" />
          Clear
        </Button>
      </div>

      {tasks.map((task, i) => (
        <Card key={task.taskID} size="sm">
          <CardContent className="flex items-center gap-3">
            <IconClock className="size-4 shrink-0 text-muted-foreground" />
            <div className="min-w-0 flex-1">
              <p className="truncate text-sm font-medium">{task.filename}</p>
              <p className="text-xs text-muted-foreground">
                {new Date(task.timestamp).toLocaleString()}
              </p>
            </div>
            <Badge
              variant={
                task.status === 'completed'
                  ? 'default'
                  : task.status === 'failed'
                    ? 'destructive'
                    : 'secondary'
              }
            >
              {task.status}
            </Badge>
            {task.status === 'completed' && (
              <Button
                variant="ghost"
                size="icon-xs"
                onClick={() => handleDownload(task.taskID)}
                disabled={downloading === task.taskID}
              >
                <IconDownload className="size-3" />
              </Button>
            )}
            <Button
              variant="ghost"
              size="icon-xs"
              onClick={() => removeTask(task.taskID)}
            >
              <IconTrash className="size-3" />
            </Button>
          </CardContent>
          {i < tasks.length - 1 && <Separator />}
        </Card>
      ))}
    </div>
  );
}

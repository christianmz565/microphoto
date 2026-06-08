import { IconClock, IconPhoto } from '@tabler/icons-react';
import { useCallback, useState } from 'react';
import { FilterSelector } from '@/components/FilterSelector';
import { ImageUploader } from '@/components/ImageUploader';
import { ProgressTracker } from '@/components/ProgressTracker';
import { ResultPreview } from '@/components/ResultPreview';
import { TaskHistory } from '@/components/TaskHistory';
import { Button } from '@/components/ui/button';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { useSSE } from '@/hooks/useSSE';
import { useTaskHistory } from '@/hooks/useTaskHistory';
import type { FilterType } from '@/lib/api';
import { getResult, uploadImage } from '@/lib/api';

type AppState = 'idle' | 'selected' | 'processing' | 'complete' | 'failed';

export default function App() {
  const [state, setState] = useState<AppState>('idle');
  const [imageFile, setImageFile] = useState<File | null>(null);
  const [taskID, setTaskID] = useState<string | null>(null);
  const [resultBlob, setResultBlob] = useState<Blob | null>(null);
  const [error, setError] = useState<string | null>(null);

  const { addTask, updateStatus } = useTaskHistory();
  const sse = useSSE(state === 'processing' ? taskID : null);

  const handleImageSelect = useCallback((file: File) => {
    setImageFile(file);
    setState('selected');
  }, []);

  const handleFilterSelect = useCallback(
    async (type: FilterType, params: Record<string, string>) => {
      if (!imageFile) return;

      setState('processing');
      setError(null);

      try {
        const id = await uploadImage(imageFile, type, params);
        setTaskID(id);
        addTask({
          taskID: id,
          filename: imageFile.name,
          filterType: type,
          status: 'processing',
        });
      } catch (err) {
        setState('failed');
        setError(err instanceof Error ? err.message : 'Upload failed');
      }
    },
    [imageFile, addTask],
  );

  // Watch SSE for completion
  if (state === 'processing' && sse.status === 'completed' && taskID) {
    const currentTaskID = taskID;
    // Fetch result
    getResult(currentTaskID)
      .then((blob) => {
        setResultBlob(blob);
        setState('complete');
        updateStatus(currentTaskID, 'completed');
      })
      .catch(() => {
        setState('failed');
        setError('Failed to download result');
        updateStatus(currentTaskID, 'failed');
      });
  }

  if (state === 'processing' && sse.status === 'failed' && taskID) {
    const currentTaskID = taskID;
    setState('failed');
    setError(sse.message || 'Processing failed');
    updateStatus(currentTaskID, 'failed');
  }

  const handleReset = useCallback(() => {
    setState('idle');
    setImageFile(null);
    setTaskID(null);
    setResultBlob(null);
    setError(null);
    sse.reset();
  }, [sse]);

  return (
    <main className="flex flex-1 flex-col items-center justify-center p-6">
      <Tabs defaultValue="upload" className="w-full max-w-md">
        <TabsList className="mb-6 w-full">
          <TabsTrigger value="upload" className="flex-1">
            <IconPhoto className="size-4" />
            Upload
          </TabsTrigger>
          <TabsTrigger value="history" className="flex-1">
            <IconClock className="size-4" />
            History
          </TabsTrigger>
        </TabsList>

        <TabsContent value="upload">
          {state === 'idle' && (
            <ImageUploader onImageSelect={handleImageSelect} />
          )}

          {state === 'selected' && (
            <FilterSelector onFilterSelect={handleFilterSelect} />
          )}

          {state === 'processing' && taskID && (
            <ProgressTracker taskID={taskID} />
          )}

          {state === 'complete' && resultBlob && (
            <ResultPreview resultBlob={resultBlob} onReset={handleReset} />
          )}

          {state === 'failed' && (
            <div className="mx-auto flex w-full max-w-md flex-col items-center gap-4">
              <p className="text-destructive">{error}</p>
              <Button onClick={handleReset}>Try again</Button>
            </div>
          )}
        </TabsContent>

        <TabsContent value="history">
          <TaskHistory />
        </TabsContent>
      </Tabs>
    </main>
  );
}

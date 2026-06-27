import { IconClock, IconPhoto } from '@tabler/icons-react';
import { useCallback, useEffect, useState } from 'react';
import { ImageEditor } from '@/components/ImageEditor';
import { ImageUploader } from '@/components/ImageUploader';
import { ProgressTracker } from '@/components/ProgressTracker';
import { ResultPreview } from '@/components/ResultPreview';
import { TaskHistory } from '@/components/TaskHistory';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { useSSE } from '@/hooks/useSSE';
import { useTaskHistory } from '@/hooks/useTaskHistory';
import type { FilterType } from '@/lib/api';
import { getResult, uploadImage } from '@/lib/api';

type AppState = 'idle' | 'editing' | 'processing' | 'complete' | 'failed';

export default function App() {
  const [state, setState] = useState<AppState>('idle');
  const [imageFile, setImageFile] = useState<File | null>(null);
  const [taskID, setTaskID] = useState<string | null>(null);
  const [resultBlob, setResultBlob] = useState<Blob | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [showFilterPicker, setShowFilterPicker] = useState(false);

  const { addTask, updateStatus } = useTaskHistory();
  const sse = useSSE(
    state === 'processing' || state === 'complete' ? taskID : null,
  );

  const handleImageSelect = useCallback((file: File) => {
    setImageFile(file);
    setState('editing');
  }, []);

  const handleSendToBackend = useCallback(() => {
    setShowFilterPicker(true);
  }, []);

  const handleFilterConfirm = useCallback(
    async (type: FilterType, params: Record<string, string>) => {
      if (!imageFile) return;

      setShowFilterPicker(false);
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
        setError(err instanceof Error ? err.message : 'Error al subir');
      }
    },
    [imageFile, addTask],
  );

  useEffect(() => {
    if (state !== 'processing' || !taskID) return;

    if (sse.status === 'JOB_COMPLETED') {
      getResult(taskID)
        .then((blob) => {
          setResultBlob(blob);
          setState('complete');
          updateStatus(taskID, 'completed');
        })
        .catch(() => {
          setState('failed');
          setError('Error al descargar resultado');
          updateStatus(taskID, 'failed');
        });
    } else if (sse.status === 'JOB_FAILED') {
      setState('failed');
      setError(sse.message || 'Error al procesar');
      updateStatus(taskID, 'failed');
    }
  }, [state, taskID, sse.status, sse.message, updateStatus]);

  const handleReset = useCallback(() => {
    setState('idle');
    setImageFile(null);
    setTaskID(null);
    setResultBlob(null);
    setError(null);
    setShowFilterPicker(false);
    sse.reset();
  }, [sse]);

  return (
    <main className="dark flex min-h-[calc(100vh-57px)] flex-col items-center justify-center p-6">
      {state !== 'editing' && (
        <Tabs defaultValue="upload" className="w-full max-w-md">
          <TabsList className="mb-6 w-full">
            <TabsTrigger value="upload" className="flex-1">
              <IconPhoto className="size-4" />
              Subir
            </TabsTrigger>
            <TabsTrigger value="history" className="flex-1">
              <IconClock className="size-4" />
              Historial
            </TabsTrigger>
          </TabsList>

          <TabsContent value="upload">
            {state === 'idle' && (
              <ImageUploader onImageSelect={handleImageSelect} />
            )}

            {state === 'processing' && taskID && (
              <ProgressTracker taskID={taskID} />
            )}

            {state === 'complete' && resultBlob && taskID && (
              <ResultPreview
                resultBlob={resultBlob}
                onReset={handleReset}
                taskID={taskID}
              />
            )}

            {state === 'failed' && (
              <div className="mx-auto flex w-full max-w-md flex-col items-center gap-4">
                <p className="text-destructive">{error}</p>
                <Button onClick={handleReset}>Intentar de nuevo</Button>
              </div>
            )}
          </TabsContent>

          <TabsContent value="history">
            <TaskHistory />
          </TabsContent>
        </Tabs>
      )}

      {state === 'editing' && imageFile && (
        <>
          <ImageEditor
            file={imageFile}
            onSendToBackend={handleSendToBackend}
            onBack={handleReset}
          />

          {showFilterPicker && (
            <FilterPickerDialog
              onConfirm={handleFilterConfirm}
              onCancel={() => setShowFilterPicker(false)}
            />
          )}
        </>
      )}
    </main>
  );
}

interface FilterPickerDialogProps {
  onConfirm: (type: FilterType, params: Record<string, string>) => void;
  onCancel: () => void;
}

const filterOptions: {
  type: FilterType;
  label: string;
  description: string;
}[] = [
  { type: 'GRAYSCALE', label: 'Escala de grises', description: 'Convertir a blanco y negro' },
  { type: 'BLUR', label: 'Desenfoque', description: 'Aplicar desenfoque gaussiano' },
  { type: 'BRIGHTNESS', label: 'Brillo', description: 'Ajustar brillo de la imagen' },
  { type: 'RESIZE', label: 'Redimensionar', description: 'Cambiar dimensiones de la imagen' },
];

function FilterPickerDialog({ onConfirm, onCancel }: FilterPickerDialogProps) {
  const [selected, setSelected] = useState<FilterType | null>(null);
  const [params, setParams] = useState<Record<string, string>>({});

  const updateParam = (key: string, value: string) => {
    setParams((prev) => ({ ...prev, [key]: value }));
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm">
      <Card className="mx-4 w-full max-w-sm border-zinc-800 bg-zinc-950">
        <CardHeader>
          <CardTitle className="text-sm font-medium text-zinc-300">
            Elegir filtro para procesar en servidor
          </CardTitle>
        </CardHeader>
        <CardContent className="flex flex-col gap-4">
          <div className="grid grid-cols-2 gap-2">
            {filterOptions.map((f) => (
              <Button
                key={f.type}
                variant={selected === f.type ? 'default' : 'outline'}
                size="sm"
                onClick={() => {
                  setSelected(f.type);
                  setParams({});
                }}
                className="justify-start text-xs"
              >
                {f.label}
              </Button>
            ))}
          </div>

          {selected === 'BLUR' && (
            <div className="flex flex-col gap-1.5">
              <label htmlFor="blur-radius" className="text-xs text-muted-foreground">
                Radius (1-100)
              </label>
              <input
                id="blur-radius"
                type="number"
                min="1"
                max="100"
                placeholder="10"
                value={params.radius ?? ''}
                onChange={(e) => updateParam('radius', e.target.value)}
                className="rounded-md border border-border bg-background px-3 py-1.5 text-sm"
              />
            </div>
          )}

          {selected === 'BRIGHTNESS' && (
            <div className="flex flex-col gap-1.5">
              <label htmlFor="brightness-factor" className="text-xs text-muted-foreground">
                Factor (0.1-3.0)
              </label>
              <input
                id="brightness-factor"
                type="number"
                min="0.1"
                max="3"
                step="0.1"
                placeholder="1.5"
                value={params.factor ?? ''}
                onChange={(e) => updateParam('factor', e.target.value)}
                className="rounded-md border border-border bg-background px-3 py-1.5 text-sm"
              />
            </div>
          )}

          {selected === 'RESIZE' && (
            <div className="flex gap-2">
              <div className="flex flex-1 flex-col gap-1.5">
                <label htmlFor="resize-width" className="text-xs text-muted-foreground">
                  width
                </label>
                <input
                  id="resize-width"
                  type="number"
                  min="1"
                  placeholder="800"
                  value={params.width ?? ''}
                  onChange={(e) => updateParam('width', e.target.value)}
                  className="rounded-md border border-border bg-background px-3 py-1.5 text-sm"
                />
              </div>
              <div className="flex flex-1 flex-col gap-1.5">
                <label htmlFor="resize-height" className="text-xs text-muted-foreground">
                  height
                </label>
                <input
                  id="resize-height"
                  type="number"
                  min="1"
                  placeholder="600"
                  value={params.height ?? ''}
                  onChange={(e) => updateParam('height', e.target.value)}
                  className="rounded-md border border-border bg-background px-3 py-1.5 text-sm"
                />
              </div>
            </div>
          )}

          <div className="flex gap-2">
            <Button variant="outline" onClick={onCancel} className="flex-1">
              Cancelar
            </Button>
            <Button
              disabled={!selected}
              onClick={() => selected && onConfirm(selected, params)}
              className="flex-1"
            >
              Procesar
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

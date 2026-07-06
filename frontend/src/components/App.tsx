import { IconClock, IconPhoto } from '@tabler/icons-react';
import { useCallback, useEffect, useRef, useState } from 'react';
import { ImageEditor } from '@/components/ImageEditor';
import { ImageUploader } from '@/components/ImageUploader';
import { TaskHistory } from '@/components/TaskHistory';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';

export default function App() {
  const [imageFile, setImageFile] = useState<File | null>(null);
  const [isDragging, setIsDragging] = useState(false);
  const dragCounter = useRef(0);

  const handleFileSelect = useCallback((file: File) => {
    setImageFile(file);
  }, []);

  const handleBack = useCallback(() => {
    setImageFile(null);
  }, []);

  const handleDragEnter = useCallback((e: DragEvent) => {
    e.preventDefault();
    dragCounter.current++;
    if (e.dataTransfer?.items && e.dataTransfer.items.length > 0) {
      setIsDragging(true);
    }
  }, []);

  const handleDragLeave = useCallback((e: DragEvent) => {
    e.preventDefault();
    dragCounter.current--;
    if (dragCounter.current === 0) {
      setIsDragging(false);
    }
  }, []);

  const handleDragOver = useCallback((e: DragEvent) => {
    e.preventDefault();
  }, []);

  const handleDrop = useCallback((e: DragEvent) => {
    e.preventDefault();
    setIsDragging(false);
    dragCounter.current = 0;

    const file = e.dataTransfer?.files?.[0];
    if (file) {
      const isImage = file.type.startsWith('image/');
      const isVideoFile = file.type.startsWith('video/');
      const MAX_SIZE = 2 * 1024 * 1024 * 1024; // 2GB
      if ((isImage || isVideoFile) && file.size <= MAX_SIZE) {
        setImageFile(file);
      }
    }
  }, []);

  useEffect(() => {
    if (imageFile) return;

    window.addEventListener('dragenter', handleDragEnter);
    window.addEventListener('dragleave', handleDragLeave);
    window.addEventListener('dragover', handleDragOver);
    window.addEventListener('drop', handleDrop);
    return () => {
      window.removeEventListener('dragenter', handleDragEnter);
      window.removeEventListener('dragleave', handleDragLeave);
      window.removeEventListener('dragover', handleDragOver);
      window.removeEventListener('drop', handleDrop);
    };
  }, [imageFile, handleDragEnter, handleDragLeave, handleDragOver, handleDrop]);

  return (
    <main className="dark flex min-h-[calc(100vh-57px)] flex-col items-center justify-center p-6">
      {isDragging && (
        <div className="fixed inset-0 z-50 flex flex-col items-center justify-center bg-black/80 backdrop-blur-md animate-in fade-in duration-200">
          <div className="flex flex-col items-center gap-4 rounded-3xl border-2 border-dashed border-zinc-700 bg-zinc-900/50 p-12 text-center max-w-md mx-6">
            <div className="flex size-16 items-center justify-center rounded-full bg-primary/10 text-primary">
              <IconPhoto className="size-8 animate-bounce" />
            </div>
            <div className="space-y-1">
              <p className="text-lg font-semibold text-zinc-100">
                Suelte el archivo aquí
              </p>
              <p className="text-sm text-zinc-400">
                Imágenes y videos de hasta 2GB para procesar
              </p>
            </div>
          </div>
        </div>
      )}

      {!imageFile && (
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
            <ImageUploader onFileSelect={handleFileSelect} />
          </TabsContent>

          <TabsContent value="history">
            <TaskHistory />
          </TabsContent>
        </Tabs>
      )}

      {imageFile && <ImageEditor file={imageFile} onBack={handleBack} />}
    </main>
  );
}

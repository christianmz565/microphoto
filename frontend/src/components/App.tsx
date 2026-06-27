import { IconClock, IconPhoto } from '@tabler/icons-react';
import { useCallback, useState } from 'react';
import { ImageEditor } from '@/components/ImageEditor';
import { ImageUploader } from '@/components/ImageUploader';
import { TaskHistory } from '@/components/TaskHistory';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';

export default function App() {
  const [imageFile, setImageFile] = useState<File | null>(null);

  const handleFileSelect = useCallback((file: File) => {
    setImageFile(file);
  }, []);

  const handleBack = useCallback(() => {
    setImageFile(null);
  }, []);

  return (
    <main className="dark flex min-h-[calc(100vh-57px)] flex-col items-center justify-center p-6">
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

      {imageFile && (
        <ImageEditor file={imageFile} onBack={handleBack} />
      )}
    </main>
  );
}

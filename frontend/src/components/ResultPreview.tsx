import { IconDownload, IconRefresh } from '@tabler/icons-react';

import { useEffect, useState } from 'react';

import { ProgressTracker } from '@/components/ProgressTracker';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';

interface ResultPreviewProps {
  resultBlob: Blob;
  onReset: () => void;
  taskID: string;
  isVideo?: boolean;
}

export function ResultPreview({
  resultBlob,
  onReset,
  taskID,
  isVideo,
}: ResultPreviewProps) {
  const [fileUrl, setFileUrl] = useState<string>('');

  useEffect(() => {
    const url = URL.createObjectURL(resultBlob);
    setFileUrl(url);
    return () => {
      URL.revokeObjectURL(url);
    };
  }, [resultBlob]);

  const handleDownload = () => {
    const a = document.createElement('a');
    a.href = fileUrl;
    a.download = isVideo ? 'processed.mp4' : 'processed.png';
    a.click();
  };

  return (
    <div className="flex flex-col gap-6">
      <Card className="mx-auto w-full max-w-md border-zinc-800 bg-black/40 backdrop-blur-xl">
        <CardHeader>
          <CardTitle className="text-sm font-medium tracking-tight text-zinc-400 uppercase">
            Resultado
          </CardTitle>
        </CardHeader>
        <CardContent className="flex flex-col items-center gap-6">
          {fileUrl &&
            (isVideo ? (
              <video
                src={fileUrl}
                className="max-h-80 rounded-lg object-contain shadow-2xl"
                controls
                autoPlay
                loop
                muted
              />
            ) : (
              <img
                src={fileUrl}
                alt="Processed"
                className="max-h-80 rounded-lg object-contain shadow-2xl"
              />
            ))}
          <div className="flex w-full gap-3">
            <Button onClick={handleDownload} className="flex-1">
              <IconDownload className="size-4" />
              Descargar
            </Button>
            <Button variant="outline" onClick={onReset} className="flex-1">
              <IconRefresh className="size-4" />
              Procesar otra
            </Button>
          </div>
        </CardContent>
      </Card>

      <ProgressTracker taskID={taskID} isVideo={isVideo} />
    </div>
  );
}

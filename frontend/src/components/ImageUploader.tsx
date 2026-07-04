import { IconPhoto, IconUpload, IconVideo } from '@tabler/icons-react';
import { useCallback, useEffect, useRef, useState } from 'react';

import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';

interface ImageUploaderProps {
  onFileSelect: (file: File) => void;
}

const ACCEPTED_TYPES = 'image/*,video/*';
const MAX_SIZE = 2 * 1024 * 1024 * 1024; // 2GB

export function ImageUploader({ onFileSelect }: ImageUploaderProps) {
  const [preview, setPreview] = useState<string | null>(null);
  const [filename, setFilename] = useState<string | null>(null);
  const [isVideo, setIsVideo] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);

  const handleFile = useCallback(
    (file: File) => {
      const isImage = file.type.startsWith('image/');
      const isVideoFile = file.type.startsWith('video/');
      if (!isImage && !isVideoFile) return;
      if (file.size > MAX_SIZE) return;

      setFilename(file.name);
      setIsVideo(isVideoFile);
      setPreview(URL.createObjectURL(file));
      onFileSelect(file);
    },
    [onFileSelect],
  );

  const handleDrop = useCallback(
    (e: React.DragEvent) => {
      e.preventDefault();
      const file = e.dataTransfer.files[0];
      if (file) handleFile(file);
    },
    [handleFile],
  );

  const handleDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault();
  }, []);

  const handleChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const file = e.target.files?.[0];
      if (file) handleFile(file);
    },
    [handleFile],
  );

  const handlePaste = useCallback(
    async (e: ClipboardEvent) => {
      const items = e.clipboardData?.items;
      if (!items) return;

      for (const item of items) {
        if (item.type.startsWith('image/')) {
          e.preventDefault();
          const blob = item.getAsFile();
          if (blob) {
            const pastedFile = new File([blob], 'pasted-image.png', {
              type: blob.type,
            });
            handleFile(pastedFile);
          }
          return;
        }
      }
    },
    [handleFile],
  );

  useEffect(() => {
    window.addEventListener('paste', handlePaste);
    return () => window.removeEventListener('paste', handlePaste);
  }, [handlePaste]);

  if (preview) {
    return (
      <Card className="mx-auto w-full max-w-md">
        <CardContent className="flex flex-col items-center gap-4">
          {isVideo ? (
            <video
              src={preview}
              className="max-h-64 rounded-2xl object-contain"
              controls
              muted
            />
          ) : (
            <img
              src={preview}
              alt="Preview"
              className="max-h-64 rounded-2xl object-contain"
            />
          )}
          <div className="flex items-center gap-2 text-sm text-muted-foreground">
            {isVideo ? (
              <IconVideo className="size-4" />
            ) : (
              <IconPhoto className="size-4" />
            )}
            {filename}
          </div>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => {
              setPreview(null);
              setFilename(null);
              setIsVideo(false);
            }}
          >
            Elegir otro archivo
          </Button>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card
      className="mx-auto w-full max-w-md cursor-pointer border-dashed transition-colors hover:bg-muted/50"
      onDrop={handleDrop}
      onDragOver={handleDragOver}
      onClick={() => inputRef.current?.click()}
    >
      <CardContent className="flex flex-col items-center gap-4 py-12">
        <div className="flex size-16 items-center justify-center rounded-full bg-muted">
          <IconUpload className="size-8 text-muted-foreground" />
        </div>
        <div className="text-center">
          <p className="font-medium">Arrastra, pega o haz clic para buscar</p>
          <p className="text-sm text-muted-foreground">
            Imágenes y videos hasta 2GB · Ctrl+V para pegar
          </p>
        </div>
        <input
          ref={inputRef}
          type="file"
          accept={ACCEPTED_TYPES}
          className="hidden"
          onChange={handleChange}
        />
      </CardContent>
    </Card>
  );
}

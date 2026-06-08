import { IconUpload } from '@tabler/icons-react';
import { useCallback, useRef, useState } from 'react';

import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';

interface ImageUploaderProps {
  onImageSelect: (file: File) => void;
}

export function ImageUploader({ onImageSelect }: ImageUploaderProps) {
  const [preview, setPreview] = useState<string | null>(null);
  const [filename, setFilename] = useState<string | null>(null);
  const inputRef = useRef<HTMLInputElement>(null);

  const handleFile = useCallback(
    (file: File) => {
      if (!file.type.startsWith('image/')) return;
      if (file.size > 50 * 1024 * 1024) return;

      setFilename(file.name);
      setPreview(URL.createObjectURL(file));
      onImageSelect(file);
    },
    [onImageSelect],
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

  if (preview) {
    return (
      <Card className="mx-auto w-full max-w-md">
        <CardContent className="flex flex-col items-center gap-4">
          <img
            src={preview}
            alt="Preview"
            className="max-h-64 rounded-2xl object-contain"
          />
          <span className="text-sm text-muted-foreground">{filename}</span>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => {
              setPreview(null);
              setFilename(null);
            }}
          >
            Elegir otra imagen
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
          <p className="font-medium">
            Arrastra una imagen aquí o haz clic para buscar
          </p>
          <p className="text-sm text-muted-foreground">
            PNG, JPG, GIF hasta 50MB
          </p>
        </div>
        <input
          ref={inputRef}
          type="file"
          accept="image/*"
          className="hidden"
          onChange={handleChange}
        />
      </CardContent>
    </Card>
  );
}

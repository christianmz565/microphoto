import {
  IconDownload,
  IconPhoto,
  IconServer,
} from '@tabler/icons-react';
import { useCallback, useEffect, useState } from 'react';

import { EditorCanvas } from '@/components/EditorCanvas';
import { EffectControls } from '@/components/EffectControls';
import { Button } from '@/components/ui/button';
import { useImageProcessor } from '@/hooks/useImageProcessor';

interface ImageEditorProps {
  file: File;
  onSendToBackend: () => void;
  onBack: () => void;
}

export function ImageEditor({ file, onSendToBackend, onBack }: ImageEditorProps) {
  const {
    loadImage,
    outputCanvasRef,
    effects,
    updateEffect,
    resetEffects,
    exportImageAsync,
    imageLoaded,
    dimensions,
  } = useImageProcessor();

  const [isExporting, setIsExporting] = useState(false);

  useEffect(() => {
    loadImage(file);
  }, [file, loadImage]);

  const handleDownload = useCallback(async () => {
    setIsExporting(true);
    try {
      const blob = await exportImageAsync('png');
      if (blob) {
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `edited-${file.name}`;
        a.click();
        URL.revokeObjectURL(url);
      }
    } finally {
      setIsExporting(false);
    }
  }, [exportImageAsync, file.name]);

  const hasChanges =
    effects.grayscale !== 0 ||
    effects.blur !== 0 ||
    effects.brightness !== 1 ||
    effects.contrast !== 1;

  return (
    <div className="image-editor">
      <div className="editor-topbar">
        <Button variant="ghost" size="sm" onClick={onBack}>
          <IconPhoto className="size-4" />
          Volver
        </Button>

        <div className="editor-topbar-actions">
          {hasChanges && (
            <Button
              variant="outline"
              size="sm"
              onClick={handleDownload}
              disabled={isExporting}
            >
              <IconDownload className="size-4" />
              {isExporting ? 'Exportando...' : 'Descargar'}
            </Button>
          )}
          <Button size="sm" onClick={onSendToBackend}>
            <IconServer className="size-4" />
            Procesar en servidor
          </Button>
        </div>
      </div>

      <div className="editor-body">
        <div className="editor-canvas-area">
          {imageLoaded ? (
            <EditorCanvas
              canvasRef={outputCanvasRef}
              width={dimensions.width}
              height={dimensions.height}
            />
          ) : (
            <div className="editor-loading">
              <div className="editor-loading-spinner" />
              <span>Cargando imagen...</span>
            </div>
          )}
        </div>

        <div className="editor-sidebar">
          <EffectControls
            effects={effects}
            onEffectChange={updateEffect}
            onReset={resetEffects}
          />
        </div>
      </div>
    </div>
  );
}

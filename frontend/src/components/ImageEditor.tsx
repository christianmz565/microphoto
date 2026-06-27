import {
  IconDownload,
  IconPhoto,
  IconLoader,
} from '@tabler/icons-react';
import { useCallback, useMemo, useState } from 'react';

import { EffectControls } from '@/components/EffectControls';
import { Button } from '@/components/ui/button';
import { useImagePreview } from '@/hooks/useImagePreview';
import { previewImage } from '@/lib/api';

interface ImageEditorProps {
  file: File;
  onBack: () => void;
}

export function ImageEditor({ file, onBack }: ImageEditorProps) {
  const {
    effects,
    updateEffect,
    resetEffects,
    previewUrl,
    isProcessing,
  } = useImagePreview(file);

  const [isDownloading, setIsDownloading] = useState(false);
  const originalUrl = useMemo(() => URL.createObjectURL(file), [file]);
  const displayUrl = previewUrl || originalUrl;

  const handleDownload = useCallback(async () => {
    setIsDownloading(true);
    try {
      const effectsList: { type: string; params: Record<string, string> }[] = [];
      if (effects.grayscale > 0) effectsList.push({ type: 'GRAYSCALE', params: {} });
      if (effects.blur > 0) effectsList.push({ type: 'BLUR', params: { radius: String(effects.blur) } });
      if (effects.brightness !== 1) effectsList.push({ type: 'BRIGHTNESS', params: { factor: String(effects.brightness) } });

      const blob = effectsList.length > 0
        ? await previewImage(file, effectsList)
        : file;

      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `edited-${file.name}`;
      a.click();
      URL.revokeObjectURL(url);
    } finally {
      setIsDownloading(false);
    }
  }, [file, effects]);

  return (
    <div className="image-editor">
      <div className="editor-topbar">
        <Button variant="ghost" size="sm" onClick={onBack}>
          <IconPhoto className="size-4" />
          Volver
        </Button>

        <div className="editor-topbar-actions">
          <Button
            size="sm"
            onClick={handleDownload}
            disabled={isDownloading || isProcessing}
          >
            {isDownloading ? (
              <IconLoader className="size-4 animate-spin" />
            ) : (
              <IconDownload className="size-4" />
            )}
            {isDownloading ? 'Descargando...' : 'Descargar'}
          </Button>
        </div>
      </div>

      <div className="editor-body">
        <div className="editor-canvas-area">
          <div className="editor-canvas-wrapper">
            <img
              src={displayUrl}
              alt="Preview"
              className="editor-canvas"
            />
            {isProcessing && (
              <div className="editor-preview-overlay">
                <IconLoader className="size-5 animate-spin" />
                <span>Procesando...</span>
              </div>
            )}
          </div>
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

import {
  IconPhoto,
  IconServer,
  IconLoader,
} from '@tabler/icons-react';
import { useMemo } from 'react';

import { EffectControls } from '@/components/EffectControls';
import { Button } from '@/components/ui/button';
import { useImagePreview } from '@/hooks/useImagePreview';

interface ImageEditorProps {
  file: File;
  onSendToBackend: () => void;
  onBack: () => void;
}

export function ImageEditor({ file, onSendToBackend, onBack }: ImageEditorProps) {
  const {
    effects,
    updateEffect,
    resetEffects,
    previewUrl,
    isProcessing,
  } = useImagePreview(file);

  const originalUrl = useMemo(() => URL.createObjectURL(file), [file]);
  const displayUrl = previewUrl || originalUrl;

  return (
    <div className="image-editor">
      <div className="editor-topbar">
        <Button variant="ghost" size="sm" onClick={onBack}>
          <IconPhoto className="size-4" />
          Volver
        </Button>

        <div className="editor-topbar-actions">
          <Button size="sm" onClick={onSendToBackend}>
            <IconServer className="size-4" />
            Procesar en servidor
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

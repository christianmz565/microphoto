import {
  IconCpu,
  IconDownload,
  IconLoader,
  IconPhoto,
  IconVideo,
} from '@tabler/icons-react';
import { useCallback, useEffect, useMemo, useState } from 'react';

import { EffectControls } from '@/components/EffectControls';
import { ProgressTracker } from '@/components/ProgressTracker';
import { ResultPreview } from '@/components/ResultPreview';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { useImagePreview } from '@/hooks/useImagePreview';
import { useSSE } from '@/hooks/useSSE';
import { useTaskHistory } from '@/hooks/useTaskHistory';
import {
  type FilterType,
  getResult,
  previewImage,
  uploadImage,
  uploadVideo,
} from '@/lib/api';

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
    isVideo,
  } = useImagePreview(file);

  const { addTask, updateStatus } = useTaskHistory();
  const [taskID, setTaskID] = useState<string | null>(null);
  const [processingState, setProcessingState] = useState<
    'editing' | 'processing' | 'completed' | 'failed'
  >('editing');
  const [resultBlob, setResultBlob] = useState<Blob | null>(null);
  const [isProcessingDistributed, setIsProcessingDistributed] = useState(false);
  const [showFilteredPreview, setShowFilteredPreview] = useState(true);

  useEffect(() => {
    if (previewUrl) {
      setShowFilteredPreview(true);
    }
  }, [previewUrl]);

  const [isDownloading, setIsDownloading] = useState(false);
  const originalUrl = useMemo(() => URL.createObjectURL(file), [file]);
  const displayUrl = previewUrl || originalUrl;

  const { status: sseStatus, error: sseError } = useSSE(taskID);

  const hasActiveEffects = useMemo(() => {
    return (
      effects.grayscale > 0 || effects.blur > 0 || effects.brightness !== 1
    );
  }, [effects]);

  const activeFilter = useMemo(() => {
    if (effects.grayscale > 0) {
      return { type: 'GRAYSCALE' as FilterType, params: {} };
    }
    if (effects.blur > 0) {
      return {
        type: 'BLUR' as FilterType,
        params: { radius: String(effects.blur) },
      };
    }
    if (effects.brightness !== 1) {
      return {
        type: 'BRIGHTNESS' as FilterType,
        params: { factor: String(effects.brightness) },
      };
    }
    return null;
  }, [effects]);

  useEffect(() => {
    if (!taskID) return;

    if (sseStatus === 'JOB_COMPLETED') {
      setProcessingState('completed');
      updateStatus(taskID, 'completed');
      getResult(taskID)
        .then((blob) => {
          setResultBlob(blob);
        })
        .catch((err) => {
          console.error('Error fetching result:', err);
          setProcessingState('failed');
          updateStatus(taskID, 'failed');
        });
    } else if (sseStatus === 'JOB_FAILED' || sseError) {
      setProcessingState('failed');
      updateStatus(taskID, 'failed');
    }
  }, [sseStatus, sseError, taskID, updateStatus]);

  const handleDownload = useCallback(async () => {
    setIsDownloading(true);
    try {
      if (isVideo) {
        const a = document.createElement('a');
        a.href = originalUrl;
        a.download = file.name;
        a.click();
        return;
      }

      const HTMLImageElementsList: {
        type: string;
        params: Record<string, string>;
      }[] = [];
      if (effects.grayscale > 0)
        HTMLImageElementsList.push({ type: 'GRAYSCALE', params: {} });
      if (effects.blur > 0)
        HTMLImageElementsList.push({
          type: 'BLUR',
          params: { radius: String(effects.blur) },
        });
      if (effects.brightness !== 1)
        HTMLImageElementsList.push({
          type: 'BRIGHTNESS',
          params: { factor: String(effects.brightness) },
        });

      const blob =
        HTMLImageElementsList.length > 0
          ? await previewImage(file, HTMLImageElementsList)
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
  }, [file, effects, isVideo, originalUrl]);

  const handleProcessDistributed = useCallback(async () => {
    if (!activeFilter) return;
    setIsProcessingDistributed(true);
    try {
      let id: string;
      if (isVideo) {
        id = await uploadVideo(file, activeFilter.type, activeFilter.params);
      } else {
        id = await uploadImage(file, activeFilter.type, activeFilter.params);
      }
      setTaskID(id);
      setProcessingState('processing');
      addTask({
        taskID: id,
        filename: file.name,
        filterType: activeFilter.type,
        status: 'processing',
        isVideo,
      });
    } catch (err) {
      console.error('Error starting distributed processing:', err);
      setProcessingState('failed');
    } finally {
      setIsProcessingDistributed(false);
    }
  }, [file, activeFilter, isVideo, addTask]);

  if (processingState === 'processing' && taskID) {
    return (
      <div className="mx-auto w-full max-w-md flex flex-col gap-4">
        <h2 className="text-center text-sm font-semibold tracking-wider text-muted-foreground uppercase mb-2">
          Procesamiento Distribuido
        </h2>
        <ProgressTracker taskID={taskID} isVideo={isVideo} />
        <Button
          variant="ghost"
          size="sm"
          onClick={() => {
            setTaskID(null);
            setProcessingState('editing');
          }}
          className="mx-auto text-zinc-500 hover:text-zinc-300"
        >
          Cancelar y volver
        </Button>
      </div>
    );
  }

  if (processingState === 'completed' && taskID && resultBlob) {
    return (
      <ResultPreview
        resultBlob={resultBlob}
        taskID={taskID}
        isVideo={isVideo}
        onReset={() => {
          setTaskID(null);
          setResultBlob(null);
          setProcessingState('editing');
        }}
      />
    );
  }

  if (processingState === 'failed') {
    return (
      <Card className="mx-auto w-full max-w-md border-destructive/20 bg-zinc-950/40 backdrop-blur-xl">
        <CardHeader>
          <CardTitle className="text-sm font-medium text-destructive uppercase">
            Procesamiento Fallido
          </CardTitle>
        </CardHeader>
        <CardContent className="flex flex-col items-center gap-6">
          <p className="text-sm text-muted-foreground text-center">
            Hubo un error al procesar el archivo en el cluster. Por favor,
            inténtalo de nuevo.
          </p>
          <Button onClick={() => setProcessingState('editing')}>
            Volver al Editor
          </Button>
        </CardContent>
      </Card>
    );
  }

  return (
    <div className="image-editor w-full max-w-6xl">
      <div className="editor-topbar">
        <Button variant="ghost" size="sm" onClick={onBack}>
          {isVideo ? (
            <IconVideo className="size-4" />
          ) : (
            <IconPhoto className="size-4" />
          )}
          Volver
        </Button>

        <div className="editor-topbar-actions flex gap-2">
          {!isVideo && (
            <Button
              variant="outline"
              size="sm"
              onClick={handleDownload}
              disabled={isDownloading || isProcessing}
            >
              {isDownloading ? (
                <IconLoader className="size-4 animate-spin" />
              ) : (
                <IconDownload className="size-4" />
              )}
              {isDownloading ? 'Descargando...' : 'Descarga Rápida'}
            </Button>
          )}
        </div>
      </div>

      <div className="editor-body">
        <div className="editor-canvas-area">
          <div className="editor-canvas-wrapper">
            {isVideo ? (
              <div className="flex flex-col w-full h-full items-center justify-center relative">
                {previewUrl && showFilteredPreview ? (
                  <video
                    src={previewUrl}
                    className="editor-canvas max-h-[50vh]"
                    controls
                    autoPlay
                    loop
                    muted
                  />
                ) : (
                  <video
                    src={originalUrl}
                    className="editor-canvas max-h-[50vh]"
                    controls
                    muted
                  />
                )}

                {isProcessing && (
                  <div className="editor-preview-overlay">
                    <IconLoader className="size-5 animate-spin" />
                    <span>Generando vista previa...</span>
                  </div>
                )}

                {previewUrl && (
                  <Button
                    variant="secondary"
                    size="sm"
                    onClick={() => setShowFilteredPreview(!showFilteredPreview)}
                    className="absolute top-4 right-4 bg-zinc-950/80 border border-zinc-800 text-zinc-300 hover:bg-zinc-900"
                  >
                    {showFilteredPreview
                      ? 'Ver Video Original'
                      : 'Ver Vista Previa'}
                  </Button>
                )}

                {!previewUrl && (
                  <div className="mt-4 text-xs text-zinc-500 bg-zinc-900/60 px-3 py-1.5 rounded-full select-none text-center">
                    Modifica los parámetros en la derecha para ver una vista
                    previa del filtro sobre el primer frame.
                  </div>
                )}
              </div>
            ) : (
              <>
                <img src={displayUrl} alt="Preview" className="editor-canvas" />
                {isProcessing && (
                  <div className="editor-preview-overlay">
                    <IconLoader className="size-5 animate-spin" />
                    <span>Procesando...</span>
                  </div>
                )}
              </>
            )}
          </div>
        </div>

        <div className="editor-sidebar flex flex-col justify-between">
          <EffectControls
            effects={effects}
            onEffectChange={updateEffect}
            onReset={resetEffects}
          />

          <div className="mt-6 px-4 pb-4">
            <Button
              className="w-full flex items-center justify-center gap-2"
              onClick={handleProcessDistributed}
              disabled={isProcessingDistributed || !hasActiveEffects}
              size="lg"
            >
              {isProcessingDistributed ? (
                <IconLoader className="size-4 animate-spin" />
              ) : (
                <IconCpu className="size-4" />
              )}
              Procesar en Cluster
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}

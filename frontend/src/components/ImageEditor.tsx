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
import { Progress } from '@/components/ui/progress';
import { buildEffectsList, useImagePreview } from '@/hooks/useImagePreview';
import { useSSE } from '@/hooks/useSSE';
import { useTaskHistory } from '@/hooks/useTaskHistory';
import { getResult, previewImage, uploadImage, uploadVideo } from '@/lib/api';

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
    isPreparing,
    firstFrameBlob,
    slideshowMetadata,
    isVideo,
  } = useImagePreview(file);

  const [firstFrameUrl, setFirstFrameUrl] = useState<string | null>(null);

  useEffect(() => {
    if (!firstFrameBlob) {
      setFirstFrameUrl(null);
      return;
    }
    const url = URL.createObjectURL(firstFrameBlob);
    setFirstFrameUrl(url);
    return () => {
      URL.revokeObjectURL(url);
    };
  }, [firstFrameBlob]);

  const [currentFrameIndex, setCurrentFrameIndex] = useState(0);

  useEffect(() => {
    if (!slideshowMetadata) return;
    const interval = setInterval(() => {
      setCurrentFrameIndex((prev) => (prev + 1) % slideshowMetadata.count);
    }, 1500);
    return () => clearInterval(interval);
  }, [slideshowMetadata]);

  const { addTask, updateStatus } = useTaskHistory();
  const [taskID, setTaskID] = useState<string | null>(null);
  const [processingState, setProcessingState] = useState<
    'editing' | 'processing' | 'completed' | 'failed'
  >('editing');
  const [resultBlob, setResultBlob] = useState<Blob | null>(null);
  const [isProcessingDistributed, setIsProcessingDistributed] = useState(false);
  const [uploadProgress, setUploadProgress] = useState<number | null>(null);
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

  useEffect(() => {
    if (!taskID) return;

    if (sseStatus === 'JOB_COMPLETED') {
      updateStatus(taskID, 'completed');
      getResult(taskID)
        .then((blob) => {
          setResultBlob(blob);
          setProcessingState('completed');
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
    const effectsList = buildEffectsList(effects);
    setIsProcessingDistributed(true);
    setUploadProgress(0);
    try {
      let id: string;
      const onProgress = (p: number) => {
        setUploadProgress(p);
      };
      if (isVideo) {
        id = await uploadVideo(file, effectsList, onProgress);
      } else {
        id = await uploadImage(file, effectsList, onProgress);
      }
      setTaskID(id);
      setProcessingState('editing'); // Note: previously set to processing, wait, let's keep 'processing' as in step 4
      setProcessingState('processing');
      addTask({
        taskID: id,
        filename: file.name,
        filterType: effectsList[0]?.type || 'UNSPECIFIED',
        status: 'processing',
        isVideo,
      });
    } catch (err) {
      console.error('Error starting distributed processing:', err);
      setProcessingState('failed');
    } finally {
      setIsProcessingDistributed(false);
      setUploadProgress(null);
    }
  }, [file, effects, isVideo, addTask]);

  if (processingState === 'processing' && taskID) {
    return (
      <div className="mx-auto w-full max-w-6xl flex flex-col gap-4">
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
                  slideshowMetadata && firstFrameUrl ? (
                    <div className="editor-canvas relative overflow-hidden rounded-2xl border border-zinc-800 bg-zinc-950">
                      <img
                        src={firstFrameUrl}
                        alt="spacer"
                        className="opacity-0 pointer-events-none w-full h-auto block"
                      />
                      <img
                        src={previewUrl}
                        alt="Preview Slideshow"
                        className="absolute top-0 left-0 h-full max-w-none transition-transform duration-300 ease-in-out"
                        style={{
                          width: `${slideshowMetadata.count * 100}%`,
                          transform: `translateX(-${(currentFrameIndex / slideshowMetadata.count) * 100}%)`,
                        }}
                      />
                    </div>
                  ) : (
                    <img
                      src={previewUrl}
                      alt="Preview"
                      className="editor-canvas max-h-[50vh] rounded-2xl object-contain"
                    />
                  )
                ) : (
                  <video
                    src={originalUrl}
                    className="editor-canvas max-h-[50vh]"
                    controls
                    muted
                  />
                )}

                {(isProcessing || isPreparing) && (
                  <div className="editor-preview-overlay">
                    <IconLoader className="size-5 animate-spin" />
                    <span>
                      {isPreparing
                        ? 'Preparando vista previa...'
                        : 'Generando vista previa...'}
                    </span>
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
                {(isProcessing || isPreparing) && (
                  <div className="editor-preview-overlay">
                    <IconLoader className="size-5 animate-spin" />
                    <span>
                      {isPreparing
                        ? 'Preparando vista previa...'
                        : 'Procesando...'}
                    </span>
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

          <div className="mt-6 px-4 pb-4 flex flex-col gap-3">
            {uploadProgress !== null && (
              <div className="w-full flex flex-col gap-1.5">
                <div className="flex justify-between text-xs text-muted-foreground px-1">
                  <span>Subiendo al servidor...</span>
                  <span className="font-mono">
                    {Math.round(uploadProgress * 100)}%
                  </span>
                </div>
                <Progress
                  value={uploadProgress * 100}
                  className="h-1 bg-zinc-900"
                />
              </div>
            )}
            <Button
              className="w-full flex items-center justify-center gap-2"
              onClick={handleProcessDistributed}
              disabled={isProcessingDistributed || isPreparing}
              size="lg"
            >
              {isProcessingDistributed ? (
                <IconLoader className="size-4 animate-spin" />
              ) : (
                <IconCpu className="size-4" />
              )}
              {isProcessingDistributed ? 'Subiendo...' : 'Procesar en Cluster'}
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}

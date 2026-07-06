// Trigger module re-parsing for Vite/Astro compilation cache
import { useCallback, useEffect, useRef, useState } from 'react';
import { type PreviewEffect, previewImage } from '@/lib/api';

export interface ImageEffects {
  grayscale: number;
  blur: number;
  brightness: number;
  contrast: number;
  resize: number;
  sepia: number;
  vignette: number;
}

export interface SlideshowMetadata {
  count: number;
  width: number;
  height: number;
}

const defaultEffects: ImageEffects = {
  grayscale: 0,
  blur: 0,
  brightness: 1,
  contrast: 1,
  resize: 1,
  sepia: 0,
  vignette: 0,
};

export function buildEffectsList(effects: ImageEffects): PreviewEffect[] {
  const list: PreviewEffect[] = [];

  if (effects.grayscale > 0) {
    list.push({ type: 'GRAYSCALE', params: {} });
  }
  if (effects.blur > 0) {
    list.push({ type: 'BLUR', params: { radius: String(effects.blur) } });
  }
  if (effects.brightness !== 1) {
    list.push({
      type: 'BRIGHTNESS',
      params: { factor: String(effects.brightness) },
    });
  }
  if (effects.contrast !== 1) {
    list.push({
      type: 'CONTRAST',
      params: { factor: String(effects.contrast) },
    });
  }
  if (effects.resize !== 1) {
    list.push({
      type: 'RESIZE',
      params: { scale: String(effects.resize) },
    });
  }
  if (effects.sepia > 0) {
    list.push({
      type: 'SEPIA',
      params: { intensity: String(effects.sepia) },
    });
  }
  if (effects.vignette > 0) {
    list.push({
      type: 'VIGNETTE',
      params: { intensity: String(effects.vignette) },
    });
  }

  return list;
}

function extractVideoSlideshow(file: File): Promise<{
  blob: Blob;
  firstFrame: Blob;
  count: number;
  width: number;
  height: number;
}> {
  return new Promise((resolve, reject) => {
    const video = document.createElement('video');
    video.src = URL.createObjectURL(file);
    video.crossOrigin = 'anonymous';
    video.muted = true;
    video.playsInline = true;

    video.onloadedmetadata = async () => {
      try {
        const duration = video.duration || 1;
        const count = 5;
        const frames: HTMLCanvasElement[] = [];

        // Generar timestamps al 10%, 30%, 50%, 70%, 90% del video
        const times = Array.from(
          { length: count },
          (_, i) => ((i * 2 + 1) / (count * 2)) * duration,
        );

        const maxFrameDim = 1280;
        let w = video.videoWidth || maxFrameDim;
        let h = video.videoHeight || Math.round((maxFrameDim * 9) / 16);
        if (w > maxFrameDim || h > maxFrameDim) {
          if (w > h) {
            h = Math.round((h * maxFrameDim) / w);
            w = maxFrameDim;
          } else {
            w = Math.round((w * maxFrameDim) / h);
            h = maxFrameDim;
          }
        }

        for (const time of times) {
          await new Promise<void>((res, rej) => {
            video.currentTime = time;
            video.onseeked = () => {
              const canvas = document.createElement('canvas');
              canvas.width = w;
              canvas.height = h;
              const ctx = canvas.getContext('2d');
              if (ctx) {
                ctx.drawImage(video, 0, 0, w, h);
                frames.push(canvas);
                res();
              } else {
                rej(new Error('Failed to get context'));
              }
            };
            video.onerror = () => {
              rej(new Error('Video seek error'));
            };
          });
        }

        // Obtener el primer fotograma individual
        const firstFrameBlob = await new Promise<Blob>((res, rej) => {
          frames[0].toBlob(
            (b) => {
              if (b) {
                res(b);
              } else {
                rej(new Error('Failed to extract first frame blob'));
              }
            },
            'image/jpeg',
            0.85,
          );
        });

        // Combinar todos los fotogramas horizontalmente
        const combinedCanvas = document.createElement('canvas');
        combinedCanvas.width = w * count;
        combinedCanvas.height = h;
        const combinedCtx = combinedCanvas.getContext('2d');
        if (combinedCtx) {
          for (let i = 0; i < count; i++) {
            combinedCtx.drawImage(frames[i], i * w, 0);
          }
          combinedCanvas.toBlob(
            (blob) => {
              URL.revokeObjectURL(video.src);
              if (blob) {
                resolve({
                  blob,
                  firstFrame: firstFrameBlob,
                  count,
                  width: w,
                  height: h,
                });
              } else {
                reject(new Error('Failed to generate combined blob'));
              }
            },
            'image/jpeg',
            0.8,
          );
        } else {
          URL.revokeObjectURL(video.src);
          reject(new Error('Failed to get combined context'));
        }
      } catch (err) {
        URL.revokeObjectURL(video.src);
        reject(err);
      }
    };

    video.onerror = () => {
      URL.revokeObjectURL(video.src);
      reject(new Error('Failed to load video metadata'));
    };
  });
}

function downscaleImage(file: File): Promise<Blob> {
  return new Promise((resolve, reject) => {
    const img = new Image();
    img.src = URL.createObjectURL(file);
    img.onload = () => {
      const maxDim = 1200;
      let w = img.width;
      let h = img.height;
      if (w > maxDim || h > maxDim) {
        if (w > h) {
          h = Math.round((h * maxDim) / w);
          w = maxDim;
        } else {
          w = Math.round((w * maxDim) / h);
          h = maxDim;
        }
      }
      const canvas = document.createElement('canvas');
      canvas.width = w;
      canvas.height = h;
      const ctx = canvas.getContext('2d');
      if (ctx) {
        ctx.drawImage(img, 0, 0, w, h);
        canvas.toBlob(
          (blob) => {
            URL.revokeObjectURL(img.src);
            if (blob) {
              resolve(blob);
            } else {
              reject(new Error('Failed to convert canvas to blob'));
            }
          },
          'image/jpeg',
          0.85,
        );
      } else {
        URL.revokeObjectURL(img.src);
        reject(new Error('Failed to get 2d context'));
      }
    };
    img.onerror = () => {
      URL.revokeObjectURL(img.src);
      reject(new Error('Failed to load image'));
    };
  });
}

export function useImagePreview(file: File) {
  const [effects, setEffects] = useState<ImageEffects>(defaultEffects);
  const [previewUrl, setPreviewUrl] = useState<string | null>(null);
  const [previewSource, setPreviewSource] = useState<Blob | null>(null);
  const [firstFrameBlob, setFirstFrameBlob] = useState<Blob | null>(null);
  const [slideshowMetadata, setSlideshowMetadata] =
    useState<SlideshowMetadata | null>(null);
  const [isPreparing, setIsPreparing] = useState(false);
  const [isProcessing, setIsProcessing] = useState(false);
  const debounceRef = useRef<ReturnType<typeof setTimeout>>(undefined);
  const abortRef = useRef<AbortController | null>(null);
  const previewIDRef = useRef<string | null>(null);

  const isVideo = file.type.startsWith('video/');

  const lastFileRef = useRef(file);
  if (lastFileRef.current !== file) {
    lastFileRef.current = file;
    previewIDRef.current = null;
  }

  useEffect(() => {
    let active = true;
    previewIDRef.current = null;
    setPreviewUrl((prev) => {
      if (prev) URL.revokeObjectURL(prev);
      return null;
    });
    setFirstFrameBlob(null);

    const prepare = async () => {
      setIsPreparing(true);
      try {
        if (file.type.startsWith('video/')) {
          const slideshow = await extractVideoSlideshow(file);
          if (active) {
            setPreviewSource(slideshow.blob);
            setFirstFrameBlob(slideshow.firstFrame);
            setSlideshowMetadata({
              count: slideshow.count,
              width: slideshow.width,
              height: slideshow.height,
            });
          }
        } else if (
          file.type.startsWith('image/') &&
          file.size > 5 * 1024 * 1024
        ) {
          const downscaled = await downscaleImage(file);
          if (active) {
            setPreviewSource(downscaled);
            setSlideshowMetadata(null);
            setFirstFrameBlob(null);
          }
        } else {
          if (active) {
            setPreviewSource(file);
            setSlideshowMetadata(null);
            setFirstFrameBlob(null);
          }
        }
      } catch (err) {
        console.error('Error preparing preview file:', err);
        if (active) {
          setPreviewSource(file);
          setSlideshowMetadata(null);
          setFirstFrameBlob(null);
        }
      } finally {
        if (active) setIsPreparing(false);
      }
    };

    prepare();

    return () => {
      active = false;
    };
  }, [file]);

  const requestPreview = useCallback(
    (currentEffects: ImageEffects) => {
      abortRef.current?.abort();
      clearTimeout(debounceRef.current);

      const effectsList = buildEffectsList(currentEffects);

      if (effectsList.length === 0) {
        setPreviewUrl(null);
        return;
      }

      debounceRef.current = setTimeout(async () => {
        const controller = new AbortController();
        abortRef.current = controller;
        setIsProcessing(true);

        try {
          const source = previewSource || file;
          const blob = await previewImage(
            source instanceof File
              ? source
              : new File([source], 'preview.jpg', { type: 'image/jpeg' }),
            effectsList,
            previewIDRef.current,
          );
          if (!controller.signal.aborted) {
            const returnedPreviewID = (blob as Blob & { previewID?: string })
              .previewID;
            if (returnedPreviewID) {
              previewIDRef.current = returnedPreviewID;
            }
            setPreviewUrl((prev) => {
              if (prev) URL.revokeObjectURL(prev);
              return URL.createObjectURL(blob);
            });
          }
        } catch (err) {
          if (!controller.signal.aborted) {
            console.error('Preview failed:', err);
          }
        } finally {
          if (!controller.signal.aborted) {
            setIsProcessing(false);
          }
        }
      }, 300);
    },
    [file, previewSource],
  );

  useEffect(() => {
    return () => {
      clearTimeout(debounceRef.current);
      abortRef.current?.abort();
    };
  }, []);

  const updateEffect = useCallback(
    (key: keyof ImageEffects, value: number) => {
      setEffects((prev) => {
        const next = { ...prev, [key]: value };
        requestPreview(next);
        return next;
      });
    },
    [requestPreview],
  );

  const resetEffects = useCallback(() => {
    setEffects(defaultEffects);
    previewIDRef.current = null;
    setPreviewUrl((prev) => {
      if (prev) URL.revokeObjectURL(prev);
      return null;
    });
    clearTimeout(debounceRef.current);
    abortRef.current?.abort();
  }, []);

  return {
    effects,
    updateEffect,
    resetEffects,
    previewUrl,
    isProcessing,
    isPreparing,
    firstFrameBlob,
    slideshowMetadata,
    isVideo,
  };
}

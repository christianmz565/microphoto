import { useCallback, useEffect, useRef, useState } from 'react';
import { previewImage, type PreviewEffect } from '@/lib/api';

export interface ImageEffects {
  grayscale: number;
  blur: number;
  brightness: number;
  contrast: number;
}

const defaultEffects: ImageEffects = {
  grayscale: 0,
  blur: 0,
  brightness: 1,
  contrast: 1,
};

function buildEffectsList(effects: ImageEffects): PreviewEffect[] {
  const list: PreviewEffect[] = [];

  if (effects.grayscale > 0) {
    list.push({ type: 'GRAYSCALE', params: {} });
  }
  if (effects.blur > 0) {
    list.push({ type: 'BLUR', params: { radius: String(effects.blur) } });
  }
  if (effects.brightness !== 1) {
    list.push({ type: 'BRIGHTNESS', params: { factor: String(effects.brightness) } });
  }

  return list;
}

export function useImagePreview(file: File) {
  const [effects, setEffects] = useState<ImageEffects>(defaultEffects);
  const [previewUrl, setPreviewUrl] = useState<string | null>(null);
  const [isProcessing, setIsProcessing] = useState(false);
  const debounceRef = useRef<ReturnType<typeof setTimeout>>(undefined);
  const abortRef = useRef<AbortController | null>(null);

  const isVideo = file.type.startsWith('video/');

  const requestPreview = useCallback(
    (currentEffects: ImageEffects) => {
      abortRef.current?.abort();
      clearTimeout(debounceRef.current);

      // Video files don't support real-time preview
      if (isVideo) {
        setPreviewUrl(null);
        return;
      }

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
          const blob = await previewImage(file, effectsList);
          if (!controller.signal.aborted) {
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
    [file, isVideo],
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
    isVideo,
  };
}

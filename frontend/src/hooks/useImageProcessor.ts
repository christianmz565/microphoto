import { useCallback, useEffect, useRef, useState } from 'react';

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

export function useImageProcessor() {
  const [effects, setEffects] = useState<ImageEffects>(defaultEffects);
  const [imageLoaded, setImageLoaded] = useState(false);
  const [dimensions, setDimensions] = useState({ width: 0, height: 0 });

  const originalImageRef = useRef<HTMLImageElement | null>(null);
  const offscreenRef = useRef<OffscreenCanvas | null>(null);
  const outputCanvasRef = useRef<HTMLCanvasElement | null>(null);
  const rafRef = useRef<number>(0);

  const loadImage = useCallback((file: File) => {
    const url = URL.createObjectURL(file);
    const img = new Image();
    img.onload = () => {
      originalImageRef.current = img;
      offscreenRef.current = new OffscreenCanvas(img.naturalWidth, img.naturalHeight);
      setDimensions({ width: img.naturalWidth, height: img.naturalHeight });
      setImageLoaded(true);
    };
    img.src = url;
  }, []);

  const loadImageFromUrl = useCallback((url: string) => {
    const img = new Image();
    img.crossOrigin = 'anonymous';
    img.onload = () => {
      originalImageRef.current = img;
      offscreenRef.current = new OffscreenCanvas(img.naturalWidth, img.naturalHeight);
      setDimensions({ width: img.naturalWidth, height: img.naturalHeight });
      setImageLoaded(true);
    };
    img.src = url;
  }, []);

  const processFrame = useCallback(() => {
    const img = originalImageRef.current;
    const offscreen = offscreenRef.current;
    const canvas = outputCanvasRef.current;
    if (!img || !offscreen || !canvas) return;

    canvas.width = img.naturalWidth;
    canvas.height = img.naturalHeight;

    const offCtx = offscreen.getContext('2d', { willReadFrequently: true });
    if (!offCtx) return;

    // Step 1: Draw with CSS blur filter (GPU-accelerated)
    offCtx.clearRect(0, 0, offscreen.width, offscreen.height);
    offCtx.filter = effects.blur > 0 ? `blur(${effects.blur}px)` : 'none';
    offCtx.drawImage(img, 0, 0);
    offCtx.filter = 'none';

    // Step 2: Pixel-level processing for grayscale, brightness, contrast
    if (effects.grayscale > 0 || effects.brightness !== 1 || effects.contrast !== 1) {
      const imageData = offCtx.getImageData(0, 0, offscreen.width, offscreen.height);
      const data = imageData.data;

      const gAmount = effects.grayscale;
      const bFactor = effects.brightness;
      const cFactor = effects.contrast;
      const cOffset = (1 - cFactor) * 127.5;

      for (let i = 0; i < data.length; i += 4) {
        let r = data[i];
        let g = data[i + 1];
        let b = data[i + 2];

        // Grayscale
        if (gAmount > 0) {
          const lum = 0.299 * r + 0.587 * g + 0.114 * b;
          r = r + (lum - r) * gAmount;
          g = g + (lum - g) * gAmount;
          b = b + (lum - b) * gAmount;
        }

        // Contrast
        if (cFactor !== 1) {
          r = r * cFactor + cOffset;
          g = g * cFactor + cOffset;
          b = b * cFactor + cOffset;
        }

        // Brightness
        if (bFactor !== 1) {
          r *= bFactor;
          g *= bFactor;
          b *= bFactor;
        }

        data[i] = Math.max(0, Math.min(255, r));
        data[i + 1] = Math.max(0, Math.min(255, g));
        data[i + 2] = Math.max(0, Math.min(255, b));
      }

      offCtx.putImageData(imageData, 0, 0);
    }

    // Step 3: Draw processed result to display canvas
    const ctx = canvas.getContext('2d');
    if (ctx) {
      ctx.clearRect(0, 0, canvas.width, canvas.height);
      ctx.drawImage(offscreen, 0, 0);
    }
  }, [effects]);

  useEffect(() => {
    if (!imageLoaded) return;
    cancelAnimationFrame(rafRef.current);
    rafRef.current = requestAnimationFrame(processFrame);
    return () => cancelAnimationFrame(rafRef.current);
  }, [imageLoaded, processFrame]);

  const updateEffect = useCallback((key: keyof ImageEffects, value: number) => {
    setEffects((prev) => ({ ...prev, [key]: value }));
  }, []);

  const resetEffects = useCallback(() => {
    setEffects(defaultEffects);
  }, []);

  const exportImageAsync = useCallback(async (format: 'png' | 'jpeg' = 'png', quality = 0.92): Promise<Blob | null> => {
    const canvas = outputCanvasRef.current;
    if (!canvas) return null;

    return new Promise((resolve) => {
      canvas.toBlob(
        (blob) => resolve(blob),
        `image/${format}`,
        quality,
      );
    });
  }, []);

  return {
    loadImage,
    loadImageFromUrl,
    outputCanvasRef,
    effects,
    updateEffect,
    resetEffects,
    exportImageAsync,
    imageLoaded,
    dimensions,
    originalImageRef,
  };
}

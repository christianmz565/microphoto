import { PUBLIC_API_URL } from 'astro:env/client';

export type FilterType = 'GRAYSCALE' | 'BLUR' | 'BRIGHTNESS' | 'RESIZE';

export interface ProcessResponse {
  task_id: string;
}

export function uploadFileWithProgress(
  endpoint: string,
  fieldName: string,
  file: File | Blob,
  effects: PreviewEffect[],
  filename?: string,
  onProgress?: (progress: number) => void,
): Promise<string> {
  return new Promise((resolve, reject) => {
    const xhr = new XMLHttpRequest();
    const formData = new FormData();
    if (filename) {
      formData.append(fieldName, file, filename);
    } else {
      formData.append(fieldName, file);
    }

    if (effects.length > 0) {
      formData.append('type', effects[0].type);
      for (const [key, value] of Object.entries(effects[0].params)) {
        if (value !== '') {
          formData.append(key, value);
        }
      }
    } else {
      formData.append('type', 'UNSPECIFIED');
    }
    formData.append('effects', JSON.stringify(effects));

    if (onProgress && xhr.upload) {
      xhr.upload.addEventListener('progress', (event) => {
        if (event.lengthComputable) {
          const progress = event.loaded / event.total;
          onProgress(progress);
        }
      });
    }

    xhr.addEventListener('load', () => {
      if (xhr.status >= 200 && xhr.status < 300) {
        try {
          const data: ProcessResponse = JSON.parse(xhr.responseText);
          resolve(data.task_id);
        } catch (_e) {
          reject(new Error('Failed to parse response'));
        }
      } else {
        reject(new Error(`Upload failed: ${xhr.status} ${xhr.statusText}`));
      }
    });

    xhr.addEventListener('error', () => {
      reject(new Error('Network error during upload'));
    });

    xhr.addEventListener('abort', () => {
      reject(new Error('Upload aborted'));
    });

    xhr.open('POST', `${PUBLIC_API_URL}${endpoint}`);
    xhr.send(formData);
  });
}

export async function uploadImage(
  file: File,
  effects: PreviewEffect[],
  onProgress?: (progress: number) => void,
): Promise<string> {
  return uploadFileWithProgress(
    '/api/v1/process',
    'image',
    file,
    effects,
    file.name,
    onProgress,
  );
}

export async function uploadVideo(
  file: File,
  effects: PreviewEffect[],
  onProgress?: (progress: number) => void,
): Promise<string> {
  return uploadFileWithProgress(
    '/api/v1/process-video',
    'video',
    file,
    effects,
    file.name,
    onProgress,
  );
}

export async function getResult(taskID: string): Promise<Blob> {
  const res = await fetch(`${PUBLIC_API_URL}/api/v1/result/${taskID}`);

  if (!res.ok) {
    throw new Error(`Failed to get result: ${res.status} ${res.statusText}`);
  }

  return res.blob();
}

export function connectToEvents(taskID: string): EventSource {
  return new EventSource(`${PUBLIC_API_URL}/api/v1/events/${taskID}`);
}

export interface PreviewEffect {
  type: string;
  params: Record<string, string>;
}

export async function previewImage(
  file: File | null,
  effects: PreviewEffect[],
  previewID?: string | null,
): Promise<Blob> {
  const formData = new FormData();
  if (previewID) {
    formData.append('preview_id', previewID);
  } else if (file) {
    formData.append('image', file);
  }
  formData.append('effects', JSON.stringify(effects));

  const res = await fetch(`${PUBLIC_API_URL}/api/v1/preview`, {
    method: 'POST',
    body: formData,
  });

  if (!res.ok) {
    throw new Error(`Preview failed: ${res.status} ${res.statusText}`);
  }

  const blob = await res.blob();
  const returnedPreviewID = res.headers.get('X-Preview-ID');
  if (returnedPreviewID) {
    Object.defineProperty(blob, 'previewID', {
      value: returnedPreviewID,
      writable: true,
      enumerable: true,
      configurable: true,
    });
  }

  return blob;
}

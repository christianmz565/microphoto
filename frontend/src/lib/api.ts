import { PUBLIC_API_URL } from 'astro:env/client';

export type FilterType = 'GRAYSCALE' | 'BLUR' | 'BRIGHTNESS' | 'RESIZE';

export interface ProcessResponse {
  task_id: string;
}

export async function uploadImage(
  file: File,
  type: FilterType,
  params: Record<string, string>,
): Promise<string> {
  const formData = new FormData();
  formData.append('image', file);
  formData.append('type', type);

  for (const [key, value] of Object.entries(params)) {
    if (value !== '') {
      formData.append(key, value);
    }
  }

  const res = await fetch(`${PUBLIC_API_URL}/api/v1/process`, {
    method: 'POST',
    body: formData,
  });

  if (!res.ok) {
    throw new Error(`Upload failed: ${res.status} ${res.statusText}`);
  }

  const data: ProcessResponse = await res.json();
  return data.task_id;
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
  file: File,
  effects: PreviewEffect[],
): Promise<Blob> {
  const formData = new FormData();
  formData.append('image', file);
  formData.append('effects', JSON.stringify(effects));

  const res = await fetch(`${PUBLIC_API_URL}/api/v1/preview`, {
    method: 'POST',
    body: formData,
  });

  if (!res.ok) {
    throw new Error(`Preview failed: ${res.status} ${res.statusText}`);
  }

  return res.blob();
}

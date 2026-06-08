import { IconDownload, IconRefresh } from '@tabler/icons-react';

import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';

interface ResultPreviewProps {
  resultBlob: Blob;
  onReset: () => void;
}

export function ResultPreview({ resultBlob, onReset }: ResultPreviewProps) {
  const imageUrl = URL.createObjectURL(resultBlob);

  const handleDownload = () => {
    const a = document.createElement('a');
    a.href = imageUrl;
    a.download = 'processed.png';
    a.click();
  };

  return (
    <Card className="mx-auto w-full max-w-md">
      <CardHeader>
        <CardTitle>Result</CardTitle>
      </CardHeader>
      <CardContent className="flex flex-col items-center gap-4">
        <img
          src={imageUrl}
          alt="Processed"
          className="max-h-80 rounded-2xl object-contain"
        />
        <div className="flex gap-2">
          <Button onClick={handleDownload}>
            <IconDownload className="size-4" />
            Download
          </Button>
          <Button variant="outline" onClick={onReset}>
            <IconRefresh className="size-4" />
            Process another
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}

import {
  IconCheck,
  IconLoader,
  IconPhoto,
  IconScissors,
  IconStack2,
} from '@tabler/icons-react';

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Progress } from '@/components/ui/progress';
import { useSSE } from '@/hooks/useSSE';

interface ProgressTrackerProps {
  taskID: string;
}

const steps = [
  { key: 'uploaded', label: 'Uploaded', icon: IconPhoto },
  { key: 'slicing', label: 'Slicing image', icon: IconScissors },
  { key: 'processing', label: 'Processing fragments', icon: IconStack2 },
  { key: 'reconstructing', label: 'Reconstructing', icon: IconStack2 },
] as const;

export function ProgressTracker({ taskID }: ProgressTrackerProps) {
  const { progress, status, message, isConnected, error } = useSSE(taskID);

  const getStepState = (stepKey: string) => {
    const order = ['uploaded', 'slicing', 'processing', 'reconstructing'];
    const currentIdx = order.indexOf(status || 'uploaded');
    const stepIdx = order.indexOf(stepKey);

    if (stepIdx < currentIdx) return 'done';
    if (stepIdx === currentIdx) return 'active';
    return 'pending';
  };

  return (
    <Card className="mx-auto w-full max-w-md">
      <CardHeader>
        <CardTitle>Processing</CardTitle>
      </CardHeader>
      <CardContent className="flex flex-col gap-4">
        <Progress value={progress * 100} />

        <div className="flex flex-col gap-3">
          {steps.map((step) => {
            const state = getStepState(step.key);
            return (
              <div key={step.key} className="flex items-center gap-3 text-sm">
                {state === 'done' ? (
                  <IconCheck className="size-4 text-green-500" />
                ) : state === 'active' ? (
                  <IconLoader className="size-4 animate-spin text-primary" />
                ) : (
                  <step.icon className="size-4 text-muted-foreground" />
                )}
                <span
                  className={state === 'pending' ? 'text-muted-foreground' : ''}
                >
                  {step.label}
                </span>
                {state === 'active' && message && (
                  <span className="ml-auto text-xs text-muted-foreground">
                    {message}
                  </span>
                )}
              </div>
            );
          })}
        </div>

        {error && <p className="text-sm text-destructive">{error}</p>}

        {!isConnected && !error && (
          <p className="text-sm text-muted-foreground">Connecting...</p>
        )}
      </CardContent>
    </Card>
  );
}

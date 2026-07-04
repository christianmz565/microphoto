import {
  IconActivity,
  IconCheck,
  IconChevronDown,
  IconChevronUp,
  IconLoader,
  IconScissors,
  IconStack2,
  IconTerminal2,
} from '@tabler/icons-react';
import { useState } from 'react';

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Progress } from '@/components/ui/progress';
import { useSSE } from '@/hooks/useSSE';
import { cn } from '@/lib/utils';

interface ProgressTrackerProps {
  taskID: string;
  isVideo?: boolean;
}

export function ProgressTracker({ taskID, isVideo }: ProgressTrackerProps) {
  const { progress, status, message, isConnected, error, workers } =
    useSSE(taskID);
  const [showIndustrial, setShowIndustrial] = useState(false);
  const [selectedWorkerId, setSelectedWorkerId] = useState<string | null>(null);

  const currentWorker = selectedWorkerId ? workers[selectedWorkerId] : null;

  const steps = isVideo
    ? ([
        { key: 'EXTRACTING', label: 'Extrayendo frames', icon: IconScissors },
        { key: 'PROCESSING', label: 'Procesando frames', icon: IconStack2 },
        { key: 'REASSEMBLING', label: 'Reensamblando video', icon: IconStack2 },
      ] as const)
    : ([
        { key: 'SLICING', label: 'Dividiendo imagen', icon: IconScissors },
        { key: 'PROCESSING', label: 'Procesando fragmentos', icon: IconStack2 },
        { key: 'RECONSTRUCTING', label: 'Reconstruyendo', icon: IconStack2 },
      ] as const);

  const getStepState = (stepKey: string) => {
    const order = isVideo
      ? ['EXTRACTING', 'PROCESSING', 'REASSEMBLING', 'JOB_COMPLETED']
      : ['SLICING', 'PROCESSING', 'RECONSTRUCTING', 'JOB_COMPLETED'];
    const currentIdx = order.indexOf(
      status || (isVideo ? 'EXTRACTING' : 'SLICING'),
    );
    const stepIdx = order.indexOf(stepKey);

    if (currentIdx === -1) return 'pending';
    if (status === 'JOB_COMPLETED') return 'done';
    if (stepIdx < currentIdx) return 'done';
    if (stepIdx === currentIdx) return 'active';
    return 'pending';
  };

  const activeWorkers = Object.values(workers).filter(
    (w) => status === 'JOB_COMPLETED' || Date.now() - w.lastUpdate < 10000,
  );

  return (
    <>
      <Card className="mx-auto w-full max-w-md border-zinc-800 bg-black/40 backdrop-blur-xl">
        <CardHeader className="flex flex-row items-center justify-between pb-2">
          <CardTitle className="text-[10px] font-bold tracking-widest text-zinc-500 uppercase">
            {status === 'JOB_COMPLETED'
              ? 'PROCESAMIENTO FINALIZADO'
              : 'PROCESAMIENTO EN CURSO'}
          </CardTitle>
          <button
            type="button"
            onClick={() => setShowIndustrial(!showIndustrial)}
            className="flex items-center gap-1 rounded-md px-2 py-1 text-[10px] font-bold tracking-widest text-zinc-500 uppercase transition-colors hover:bg-zinc-800 hover:text-zinc-300"
          >
            {showIndustrial ? 'VISTA SIMPLE' : 'DETALLES'}
            {showIndustrial ? (
              <IconChevronUp size={12} />
            ) : (
              <IconChevronDown size={12} />
            )}
          </button>
        </CardHeader>

        <CardContent className="flex flex-col gap-6">
          <div className="space-y-2">
            <div className="flex justify-between text-[10px] font-mono text-zinc-500 tabular-nums">
              <span>{status || 'INICIALIZANDO'}</span>
              <span>{Math.round(progress * 100)}%</span>
            </div>
            <Progress value={progress * 100} className="h-1 bg-zinc-900" />
          </div>

          <div className="flex flex-col gap-4">
            {steps.map((step) => {
              const state = getStepState(step.key);
              return (
                <div
                  key={step.key}
                  className={cn(
                    'flex items-center gap-3 text-xs transition-opacity duration-300',
                    state === 'pending' && 'opacity-40',
                  )}
                >
                  <div className="relative flex h-4 w-4 items-center justify-center">
                    {state === 'done' ? (
                      <IconCheck className="size-3 text-primary" stroke={3} />
                    ) : state === 'active' && status !== 'JOB_COMPLETED' ? (
                      <IconLoader className="size-3 animate-spin text-primary" />
                    ) : (
                      <step.icon className="size-3 text-zinc-500" />
                    )}
                  </div>
                  <span
                    className={cn(
                      'font-medium tracking-tight',
                      state === 'active' ? 'text-zinc-100' : 'text-zinc-400',
                    )}
                  >
                    {step.label}
                  </span>

                  {state === 'active' && !showIndustrial && (
                    <span className="ml-auto text-[10px] font-mono text-zinc-500 animate-in fade-in slide-in-from-right-1">
                      {message}
                    </span>
                  )}
                </div>
              );
            })}
          </div>

          {showIndustrial && (
            <div className="mt-2 space-y-4 border-t border-zinc-800 pt-4 animate-in fade-in slide-in-from-top-2">
              <div className="space-y-3">
                <div className="flex items-center gap-2 text-[10px] font-bold tracking-wider text-zinc-500 uppercase">
                  <IconActivity size={12} />
                  NODOS ACTIVOS
                </div>

                <div className="grid grid-cols-1 gap-2 sm:grid-cols-2">
                  {activeWorkers.map((worker) => (
                    <button
                      key={worker.id}
                      type="button"
                      onClick={() => setSelectedWorkerId(worker.id)}
                      className="flex flex-col rounded border border-zinc-800 bg-black/40 p-2 text-left font-mono text-[9px] transition-all hover:border-primary/50 hover:bg-zinc-900/50 active:scale-[0.98]"
                    >
                      <div className="mb-2 flex items-center justify-between border-b border-zinc-800/50 pb-1.5">
                        <div className="flex items-center gap-2">
                          <div className="relative h-1.5 w-1.5">
                            {status !== 'JOB_COMPLETED' && (
                              <div className="absolute inset-0 animate-ping rounded-full bg-primary/40" />
                            )}
                            <div
                              className={cn(
                                'absolute inset-0 rounded-full',
                                status === 'JOB_COMPLETED'
                                  ? 'bg-zinc-700'
                                  : 'bg-primary',
                              )}
                            />
                          </div>
                          <span className="font-bold text-zinc-300 uppercase">
                            NODE-{worker.id.slice(0, 4)}
                          </span>
                        </div>
                        <span className="text-[8px] text-zinc-600">
                          {worker.status}
                        </span>
                      </div>

                      <div className="flex flex-col gap-1 text-zinc-500">
                        {worker.logs.length > 0 ? (
                          worker.logs.slice(-3).map((log) => (
                            <div key={log.id} className="flex gap-2">
                              <span className="shrink-0 text-zinc-700">›</span>
                              <span className="truncate">{log.message}</span>
                            </div>
                          ))
                        ) : (
                          <span className="italic opacity-50 text-[8px]">
                            SIN REGISTROS
                          </span>
                        )}
                      </div>
                    </button>
                  ))}

                  {activeWorkers.length === 0 && (
                    <div className="col-span-full rounded border border-zinc-800/50 border-dashed p-4 text-center">
                      <span className="text-[10px] font-bold tracking-widest text-zinc-700 uppercase">
                        ESPERANDO SEÑAL DE TRABAJADORES...
                      </span>
                    </div>
                  )}
                </div>
              </div>

              <div className="rounded border border-zinc-800 bg-black/50 p-2 font-mono text-[9px] leading-relaxed text-zinc-500">
                <div className="flex gap-2">
                  <span className="text-primary/70 font-bold">SYSTEM:</span>
                  <span className="truncate">{message || 'EN ESPERA...'}</span>
                </div>
              </div>
            </div>
          )}

          {error && (
            <div className="mt-2 border-t border-destructive/20 pt-2 text-center text-[10px] font-medium text-destructive">
              {error}
            </div>
          )}

          {!isConnected && !error && status !== 'JOB_COMPLETED' && (
            <div className="flex items-center justify-center gap-2 py-1 text-[9px] font-bold tracking-widest text-zinc-600 uppercase">
              <IconLoader size={10} className="animate-spin" />
              CONEXIÓN PERDIDA - REINTENTANDO
            </div>
          )}
        </CardContent>
      </Card>

      <Dialog
        open={!!selectedWorkerId}
        onOpenChange={(open) => !open && setSelectedWorkerId(null)}
      >
        <DialogContent className="max-w-2xl border-zinc-800 bg-zinc-950 p-0 text-zinc-400">
          <DialogHeader className="border-b border-zinc-800 p-4">
            <DialogTitle className="flex items-center gap-2 font-mono text-sm font-bold tracking-widest text-zinc-200 uppercase">
              <IconTerminal2 className="size-4 text-primary" />
              LOGS DEL NODO: {currentWorker?.id}
            </DialogTitle>
          </DialogHeader>
          <div className="max-h-[60vh] overflow-y-auto p-4 font-mono text-xs leading-relaxed selection:bg-primary/20 selection:text-primary">
            {currentWorker?.logs.map((log, i) => (
              <div
                key={log.id}
                className="group flex gap-3 border-b border-zinc-900/50 py-1 last:border-0 animate-in fade-in slide-in-from-left-1"
              >
                <span className="shrink-0 select-none text-zinc-700">
                  {String(i + 1).padStart(3, '0')}
                </span>
                <span className="text-zinc-400 transition-colors group-hover:text-zinc-200">
                  {log.message}
                </span>
              </div>
            ))}
            {(!currentWorker || currentWorker.logs.length === 0) && (
              <div className="py-8 text-center italic opacity-30">
                NO HAY REGISTROS DISPONIBLES PARA ESTE NODO
              </div>
            )}
          </div>
        </DialogContent>
      </Dialog>
    </>
  );
}

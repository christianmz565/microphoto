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

const formatTime = (ts: number) => {
  let ms = ts;
  if (ts > 99999999999999) {
    ms = Math.floor(ts / 1000000);
  } else if (ts < 9999999999) {
    ms = ts * 1000;
  }
  const date = new Date(ms);
  return date.toLocaleTimeString([], {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: false,
  });
};

function WorkerWorkChart({
  chartData = [],
}: {
  chartData?: Array<Record<string, number | string>>;
}) {
  if (!chartData || chartData.length === 0) return null;

  const workerKeys = Array.from(
    new Set(
      chartData.flatMap((d) => Object.keys(d).filter((k) => k !== 'time')),
    ),
  );

  if (workerKeys.length === 0) {
    return (
      <div className="py-8 text-center italic opacity-35 text-[10px] text-zinc-500 font-mono">
        NO HAY SUFICIENTES DATOS PARA EL GRÁFICO
      </div>
    );
  }

  let maxVal = 1;
  for (const d of chartData) {
    for (const key of workerKeys) {
      const val = Number(d[key]) || 0;
      if (val > maxVal) maxVal = val;
    }
  }

  const paddingX = 40;
  const paddingY = 30;
  const width = 600;
  const height = 240;

  const chartWidth = width - paddingX * 2;
  const chartHeight = height - paddingY * 2;

  const colors = [
    '#3b82f6',
    '#10b981',
    '#f59e0b',
    '#ec4899',
    '#8b5cf6',
    '#06b6d4',
    '#84cc16',
    '#f43f5e',
    '#a855f7',
    '#6366f1',
  ];
  const getColor = (idx: number) => colors[idx % colors.length];

  const pointsCount = chartData.length;
  const getX = (idx: number) =>
    paddingX + (idx / Math.max(1, pointsCount - 1)) * chartWidth;
  const getY = (val: number) =>
    height - paddingY - (val / maxVal) * chartHeight;

  const yTicks = Array.from({ length: 4 }, (_, i) =>
    Math.round((i * maxVal) / 3),
  );

  const xTicksIndices =
    pointsCount > 2
      ? [0, Math.floor(pointsCount / 2), pointsCount - 1]
      : pointsCount === 2
        ? [0, 1]
        : [0];

  return (
    <div className="w-full space-y-4 border-t border-zinc-800 pt-4 font-mono">
      <div className="flex items-center gap-2 text-[10px] font-bold tracking-wider text-zinc-500 uppercase">
        <IconActivity size={12} className="text-primary" />
        RENDIMIENTO DE TRABAJO POR NODO
      </div>

      <div className="relative rounded border border-zinc-800 bg-black/60 p-3">
        <svg viewBox={`0 0 ${width} ${height}`} className="w-full h-auto">
          <title>Rendimiento de trabajo por nodo</title>
          {yTicks.map((tick, i) => {
            const y = getY(tick);
            return (
              <g
                // biome-ignore lint/suspicious/noArrayIndexKey: static ticks
                key={i}
              >
                <line
                  x1={paddingX}
                  y1={y}
                  x2={width - paddingX}
                  y2={y}
                  stroke="#1f2937"
                  strokeWidth={1}
                  strokeDasharray="4 4"
                />
                <text
                  x={paddingX - 8}
                  y={y + 3}
                  fill="#9ca3af"
                  fontSize="8"
                  textAnchor="end"
                  className="font-mono tabular-nums opacity-60"
                >
                  {tick}
                </text>
              </g>
            );
          })}

          {xTicksIndices.map((idx, i) => {
            const x = getX(idx);
            const dataPoint = chartData[idx];
            if (!dataPoint) return null;
            return (
              <text
                // biome-ignore lint/suspicious/noArrayIndexKey: static ticks
                key={i}
                x={x}
                y={height - paddingY + 16}
                fill="#9ca3af"
                fontSize="8"
                textAnchor="middle"
                className="font-mono opacity-60"
              >
                {dataPoint.time}
              </text>
            );
          })}

          {workerKeys.map((key, wIdx) => {
            const linePoints = chartData
              .map((d, idx) => {
                const val = Number(d[key]) || 0;
                return `${getX(idx)},${getY(val)}`;
              })
              .join(' ');

            return (
              <polyline
                key={key}
                fill="none"
                stroke={getColor(wIdx)}
                strokeWidth={2}
                points={linePoints}
                className="transition-all duration-300"
              />
            );
          })}
        </svg>

        <div className="mt-4 flex flex-wrap gap-x-3 gap-y-1.5 border-t border-zinc-800/50 pt-3">
          {workerKeys.map((key, wIdx) => (
            <div
              key={key}
              className="flex items-center gap-1.5 text-[8px] font-bold text-zinc-400"
            >
              <span
                className="h-2 w-2 rounded-full shrink-0"
                style={{ backgroundColor: getColor(wIdx) }}
              />
              <span className="uppercase">{key}</span>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

export function ProgressTracker({ taskID, isVideo }: ProgressTrackerProps) {
  const {
    progress,
    status,
    message,
    isConnected,
    error,
    workers,
    chartData = [],
  } = useSSE(taskID);
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

  const activeWorkers = Object.values(workers);

  return (
    <>
      <Card className="mx-auto w-full max-w-6xl border-zinc-800 bg-black/40 backdrop-blur-xl">
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

                <div className="grid grid-cols-1 gap-2 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5">
                  {activeWorkers.map((worker) => {
                    const isIdle =
                      status !== 'JOB_COMPLETED' &&
                      Date.now() - worker.lastUpdate >= 10000;
                    return (
                      <button
                        key={worker.id}
                        type="button"
                        onClick={() => setSelectedWorkerId(worker.id)}
                        className={cn(
                          'flex flex-col rounded border border-zinc-800 bg-black/40 p-2 text-left font-mono text-[9px] transition-all hover:border-primary/50 hover:bg-zinc-900/50 active:scale-[0.98]',
                          isIdle && 'opacity-50',
                        )}
                      >
                        <div className="mb-2 flex items-center justify-between border-b border-zinc-800/50 pb-1.5">
                          <div className="flex items-center gap-2">
                            <div className="relative h-1.5 w-1.5">
                              {status !== 'JOB_COMPLETED' && !isIdle && (
                                <div className="absolute inset-0 animate-ping rounded-full bg-primary/40" />
                              )}
                              <div
                                className={cn(
                                  'absolute inset-0 rounded-full',
                                  status === 'JOB_COMPLETED'
                                    ? 'bg-zinc-700'
                                    : isIdle
                                      ? 'bg-zinc-500'
                                      : 'bg-primary',
                                )}
                              />
                            </div>
                            <span className="font-bold text-zinc-300 uppercase">
                              NODE-{worker.id}
                            </span>
                          </div>
                          <span className="text-[8px] text-zinc-600">
                            {isIdle ? 'IDLE' : worker.status}
                          </span>
                        </div>

                        <div className="flex flex-col gap-1 text-zinc-500">
                          {worker.logs.length > 0 ? (
                            worker.logs.slice(-3).map((log) => (
                              <div
                                key={log.id}
                                className="flex gap-1.5 truncate"
                              >
                                <span className="shrink-0 text-zinc-700">
                                  [{formatTime(log.timestamp)}]
                                </span>
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
                    );
                  })}

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

          {status === 'JOB_COMPLETED' && (
            <WorkerWorkChart chartData={chartData} />
          )}
        </CardContent>
      </Card>

      <Dialog
        open={!!selectedWorkerId}
        onOpenChange={(open) => !open && setSelectedWorkerId(null)}
      >
        <DialogContent className="max-w-5xl border-zinc-800 bg-zinc-950 p-0 text-zinc-400">
          <DialogHeader className="border-b border-zinc-800 p-4">
            <DialogTitle className="flex items-center gap-2 font-mono text-sm font-bold tracking-widest text-zinc-200 uppercase">
              <IconTerminal2 className="size-4 text-primary" />
              LOGS DEL NODO: {currentWorker?.id}
            </DialogTitle>
          </DialogHeader>
          <div className="max-h-[60vh] overflow-y-auto p-4 font-mono text-xs leading-relaxed selection:bg-primary/20 selection:text-primary">
            {currentWorker?.logs.map((log) => (
              <div
                key={log.id}
                className="group flex gap-3 border-b border-zinc-900/50 py-1 last:border-0 animate-in fade-in slide-in-from-left-1"
              >
                <span className="shrink-0 select-none text-zinc-700">
                  [{formatTime(log.timestamp)}]
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

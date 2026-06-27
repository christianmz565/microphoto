import {
  IconAdjustments,
  IconBrightness,
  IconContrast,
  IconFilter,
  IconRefresh,
} from '@tabler/icons-react';

import type { ImageEffects } from '@/hooks/useImageProcessor';

interface EffectControlsProps {
  effects: ImageEffects;
  onEffectChange: (key: keyof ImageEffects, value: number) => void;
  onReset: () => void;
}

interface SliderConfig {
  key: keyof ImageEffects;
  label: string;
  icon: typeof IconAdjustments;
  min: number;
  max: number;
  step: number;
  default: number;
  unit: string;
}

const sliders: SliderConfig[] = [
  {
    key: 'grayscale',
    label: 'Escala de grises',
    icon: IconAdjustments,
    min: 0,
    max: 1,
    step: 0.01,
    default: 0,
    unit: '%',
  },
  {
    key: 'blur',
    label: 'Desenfoque',
    icon: IconFilter,
    min: 0,
    max: 20,
    step: 0.5,
    default: 0,
    unit: 'px',
  },
  {
    key: 'brightness',
    label: 'Brillo',
    icon: IconBrightness,
    min: 0,
    max: 3,
    step: 0.05,
    default: 1,
    unit: 'x',
  },
  {
    key: 'contrast',
    label: 'Contraste',
    icon: IconContrast,
    min: 0,
    max: 3,
    step: 0.05,
    default: 1,
    unit: 'x',
  },
];

function formatValue(value: number, unit: string): string {
  if (unit === '%') return `${Math.round(value * 100)}%`;
  if (unit === 'x') return `${value.toFixed(2)}x`;
  return `${value}${unit}`;
}

export function EffectControls({
  effects,
  onEffectChange,
  onReset,
}: EffectControlsProps) {
  return (
    <div className="effect-controls">
      <div className="effect-controls-header">
        <h3 className="effect-controls-title">Efectos</h3>
        <button type="button" className="effect-reset-btn" onClick={onReset}>
          <IconRefresh className="size-3.5" />
          Reset
        </button>
      </div>

      <div className="effect-sliders">
        {sliders.map((slider) => {
          const value = effects[slider.key];
          const isDefault = value === slider.default;
          const pct =
            ((value - slider.min) / (slider.max - slider.min)) * 100;

          return (
            <div
              key={slider.key}
              className={`effect-slider-group ${isDefault ? '' : 'active'}`}
            >
              <div className="effect-slider-header">
                <div className="effect-slider-label">
                  <slider.icon className="size-3.5" />
                  <span>{slider.label}</span>
                </div>
                <span className="effect-slider-value">
                  {formatValue(value, slider.unit)}
                </span>
              </div>

              <div className="effect-slider-track-wrapper">
                <input
                  type="range"
                  min={slider.min}
                  max={slider.max}
                  step={slider.step}
                  value={value}
                  onChange={(e) =>
                    onEffectChange(slider.key, Number.parseFloat(e.target.value))
                  }
                  className="effect-slider"
                  style={
                    {
                      '--slider-pct': `${pct}%`,
                    } as React.CSSProperties
                  }
                />
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}

import {
  IconAdjustments,
  IconBlur,
  IconSun,
  IconZoom,
} from '@tabler/icons-react';
import { useState } from 'react';

import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import type { FilterType } from '@/lib/api';

interface FilterSelectorProps {
  onFilterSelect: (type: FilterType, params: Record<string, string>) => void;
}

const filters: {
  type: FilterType;
  label: string;
  icon: typeof IconAdjustments;
}[] = [
  { type: 'GRAYSCALE', label: 'Escala de grises', icon: IconAdjustments },
  { type: 'BLUR', label: 'Desenfoque', icon: IconBlur },
  { type: 'BRIGHTNESS', label: 'Brillo', icon: IconSun },
  { type: 'RESIZE', label: 'Redimensionar', icon: IconZoom },
];

export function FilterSelector({ onFilterSelect }: FilterSelectorProps) {
  const [selected, setSelected] = useState<FilterType | null>(null);
  const [params, setParams] = useState<Record<string, string>>({});

  const handleSelect = (type: FilterType) => {
    setSelected(type);
    setParams({});
  };

  const handleProcess = () => {
    if (selected) {
      onFilterSelect(selected, params);
    }
  };

  const updateParam = (key: string, value: string) => {
    setParams((prev) => ({ ...prev, [key]: value }));
  };

  return (
    <Card className="mx-auto w-full max-w-md">
      <CardHeader>
        <CardTitle>Seleccionar filtro</CardTitle>
      </CardHeader>
      <CardContent className="flex flex-col gap-4">
        <div className="grid grid-cols-2 gap-2">
          {filters.map((f) => (
            <Button
              key={f.type}
              variant={selected === f.type ? 'default' : 'outline'}
              size="lg"
              onClick={() => handleSelect(f.type)}
            >
              <f.icon className="size-4" />
              {f.label}
            </Button>
          ))}
        </div>

        {selected === 'BLUR' && (
          <div className="flex flex-col gap-1.5">
            <label
              htmlFor="blur-radius"
              className="text-sm text-muted-foreground"
            >
              Radius (1-100)
            </label>
            <Input
              id="blur-radius"
              type="number"
              min="1"
              max="100"
              placeholder="10"
              value={params.radius ?? ''}
              onChange={(e) => updateParam('radius', e.target.value)}
            />
          </div>
        )}

        {selected === 'BRIGHTNESS' && (
          <div className="flex flex-col gap-1.5">
            <label
              htmlFor="brightness-factor"
              className="text-sm text-muted-foreground"
            >
              Factor (0.1-3.0)
            </label>
            <Input
              id="brightness-factor"
              type="number"
              min="0.1"
              max="3"
              step="0.1"
              placeholder="1.5"
              value={params.factor ?? ''}
              onChange={(e) => updateParam('factor', e.target.value)}
            />
          </div>
        )}

        {selected === 'RESIZE' && (
          <div className="flex gap-2">
            <div className="flex flex-1 flex-col gap-1.5">
              <label
                htmlFor="resize-width"
                className="text-sm text-muted-foreground"
              >
                width
              </label>
              <Input
                id="resize-width"
                type="number"
                min="1"
                placeholder="800"
                value={params.width ?? ''}
                onChange={(e) => updateParam('width', e.target.value)}
              />
            </div>
            <div className="flex flex-1 flex-col gap-1.5">
              <label
                htmlFor="resize-height"
                className="text-sm text-muted-foreground"
              >
                height
              </label>
              <Input
                id="resize-height"
                type="number"
                min="1"
                placeholder="600"
                value={params.height ?? ''}
                onChange={(e) => updateParam('height', e.target.value)}
              />
            </div>
          </div>
        )}

        <Button size="lg" disabled={!selected} onClick={handleProcess}>
          Procesar
        </Button>
      </CardContent>
    </Card>
  );
}

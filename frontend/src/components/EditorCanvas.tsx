import type { RefObject } from 'react';

interface EditorCanvasProps {
  canvasRef: RefObject<HTMLCanvasElement | null>;
  width: number;
  height: number;
}

export function EditorCanvas({ canvasRef, width, height }: EditorCanvasProps) {
  return (
    <div className="editor-canvas-wrapper">
      <canvas
        ref={canvasRef}
        width={width}
        height={height}
        className="editor-canvas"
      />
    </div>
  );
}

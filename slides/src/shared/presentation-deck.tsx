import { Deck } from "@revealjs/react";
import type { ComponentProps } from "react";
import "reveal.js/reveal.css";
import "reveal.js/theme/black.css";

export function PresentationDeck({
  config,
  children,
  ...props
}: ComponentProps<typeof Deck>) {
  return (
    <Deck
      config={{
        width: 1280,
        height: 720,
        margin: 0.04,
        minScale: 0.1,
        maxScale: 2.0,
        hash: true,
        controls: true,
        progress: true,
        center: false,
        ...config,
      }}
      {...props}
    >
      {children}
    </Deck>
  );
}

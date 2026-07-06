import { Slide } from "@revealjs/react";
import { SlideWrap } from "./slide-wrap";

export function ThanksSlide({
  color,
  variant = "default",
}: {
  color: string;
  variant?: "default" | "decorated";
}) {
  return (
    <Slide className="h-full">
      <SlideWrap
        color={color}
        tag=""
        variant={variant}
        className="justify-center items-center flex flex-col"
      >
        <h1
          className="text-8xl! font-light tracking-tight relative z-10"
          style={
            variant === "decorated"
              ? { textShadow: `0 0 30px ${color}40, 0 0 60px ${color}20` }
              : {}
          }
        >
          ¡<span style={{ color }}>Gracias</span>!
        </h1>
      </SlideWrap>
    </Slide>
  );
}

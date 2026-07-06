import { createFileRoute, redirect } from "@tanstack/react-router";

export const Route = createFileRoute("/presentations/")({
  beforeLoad: () => {
    throw redirect({
      to: "/",
      replace: true,
    });
  },
});

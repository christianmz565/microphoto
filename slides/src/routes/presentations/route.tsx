import { createFileRoute, Outlet } from "@tanstack/react-router";

export const Route = createFileRoute("/presentations")({
  validateSearch: (search: Record<string, unknown>) => search,
  component: PresentationsLayoutComponent,
});

function PresentationsLayoutComponent() {
  return (
    <div className="w-screen h-screen relative ">
      <Outlet />
    </div>
  );
}

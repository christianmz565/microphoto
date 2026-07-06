import { createFileRoute, Link } from "@tanstack/react-router";
import { presentations } from "@/lib/presentations";

export const Route = createFileRoute("/")({
  component: Index,
});

function Index() {
  return (
    <div className="flex min-h-screen flex-col items-center bg-linear-to-br from-slate-900 via-indigo-950 to-slate-900 px-4 py-12 font-sans text-slate-200">
      <header className="mb-12 text-center">
        <h1 className="mb-2 bg-linear-to-r from-indigo-400 to-purple-400 bg-clip-text text-5xl font-extrabold text-transparent">
          SD Trabajo Final - Microphoto
        </h1>
      </header>

      <div className="grid w-full max-w-3xl gap-6 sm:grid-cols-2 lg:grid-cols-3">
        {presentations.map((p) => (
          <Link
            key={p.path}
            to={p.path}
            className="group rounded-xl border border-white/10 bg-white/5 p-6 text-slate-200 no-underline transition-all duration-200 hover:cursor-pointer hover:bg-white/10 hover:ring-1 hover:ring-white/20"
          >
            <h3 className="mb-2 text-lg font-semibold">{p.title}</h3>
            <p className="text-sm text-slate-400">{p.description}</p>
          </Link>
        ))}
      </div>
    </div>
  );
}

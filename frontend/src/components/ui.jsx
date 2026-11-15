import { BackIcon } from "./icons.jsx";

const buttonVariants = {
  primary: "bg-sky-500 text-white hover:bg-sky-400 shadow-sm shadow-sky-500/20",
  secondary: "bg-slate-800 text-slate-100 hover:bg-slate-700 ring-1 ring-inset ring-slate-700",
  ghost: "text-slate-300 hover:bg-slate-800 hover:text-white",
};

export function Button({ variant = "secondary", className = "", children, ...props }) {
  return (
    <button
      type="button"
      className={`inline-flex items-center justify-center gap-2 rounded-lg px-3.5 py-2 text-sm font-medium transition
        focus:outline-none focus-visible:ring-2 focus-visible:ring-sky-400
        disabled:cursor-not-allowed disabled:opacity-40 ${buttonVariants[variant]} ${className}`}
      {...props}
    >
      {children}
    </button>
  );
}

export function Spinner({ className = "h-4 w-4" }) {
  return (
    <span
      className={`inline-block animate-spin rounded-full border-2 border-current border-r-transparent ${className}`}
      aria-label="Loading"
    />
  );
}

export function Panel({ className = "", children }) {
  return (
    <section className={`rounded-xl border border-slate-800 bg-slate-900/60 ${className}`}>{children}</section>
  );
}

/** Top bar shared by both modes: back button, title/subtitle and right-hand actions. */
export function PageHeader({ title, subtitle, onBack, actions }) {
  return (
    <header className="flex items-center gap-3 border-b border-slate-800 px-6 py-4">
      <Button variant="ghost" onClick={onBack} className="-ml-2 px-2" title="Back to start">
        <BackIcon />
      </Button>
      <div className="min-w-0 flex-1">
        <h1 className="text-lg font-semibold text-white">{title}</h1>
        {subtitle && <div className="truncate text-sm text-slate-400">{subtitle}</div>}
      </div>
      {actions}
    </header>
  );
}

/** A labelled folder path with a "Change" action, used for shared and save folders. */
export function FolderField({ label, path, onChange, children }) {
  return (
    <div className="flex min-w-0 items-center gap-3">
      <div className="min-w-0 flex-1">
        <div className="text-xs font-medium uppercase tracking-wide text-slate-500">{label}</div>
        <div className="truncate font-mono text-sm text-slate-200" title={path}>
          {path || "—"}
        </div>
      </div>
      {children}
      {onChange && (
        <Button onClick={onChange} className="shrink-0">
          Change
        </Button>
      )}
    </div>
  );
}

export function EmptyState({ icon, title, children }) {
  return (
    <div className="flex h-full flex-col items-center justify-center gap-3 p-8 text-center">
      {icon && <div className="text-slate-600">{icon}</div>}
      <div className="font-medium text-slate-300">{title}</div>
      {children && <div className="max-w-sm text-sm text-slate-500">{children}</div>}
    </div>
  );
}

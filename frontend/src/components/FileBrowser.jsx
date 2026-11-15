import { useEffect, useState } from "react";
import { dirname, joinPath } from "../lib/useFileBrowser.js";
import FileList from "./FileList.jsx";
import { FolderIcon, RefreshIcon, SearchIcon } from "./icons.jsx";
import { Button, EmptyState, Panel, Spinner } from "./ui.jsx";

function Breadcrumbs({ rootLabel, path, onNavigate }) {
  const parts = path ? path.split("/") : [];
  const crumb = "truncate rounded px-1.5 py-0.5 hover:bg-slate-800 hover:text-white";
  return (
    <nav className="flex min-w-0 flex-1 items-center gap-0.5 text-sm text-slate-400" aria-label="Folder path">
      <button type="button" onClick={() => onNavigate("")} className={`${crumb} shrink-0`}>
        {rootLabel}
      </button>
      {parts.map((part, i) => {
        const last = i === parts.length - 1;
        return (
          <span key={i} className="flex min-w-0 items-center gap-0.5">
            <span className="text-slate-600">/</span>
            {last ? (
              <span className="truncate px-1.5 font-medium text-slate-200">{part}</span>
            ) : (
              <button type="button" onClick={() => onNavigate(parts.slice(0, i + 1).join("/"))} className={crumb}>
                {part}
              </button>
            )}
          </span>
        );
      })}
    </nav>
  );
}

/**
 * A panel with breadcrumbs, a filter box, refresh and the file list for one folder tree.
 * `browser` comes from useFileBrowser. Files are identified by their full relative path.
 * `placeholder` replaces the list entirely (e.g. "not connected").
 */
function FileBrowser({ rootLabel, browser, selected, onSelect, onOpenFile, placeholder, className = "" }) {
  const { path, files, loading, open, refresh } = browser;
  const [filter, setFilter] = useState("");

  useEffect(() => setFilter(""), [path]);

  const query = filter.trim().toLowerCase();
  const visible = query ? files.filter((f) => f.name.toLowerCase().includes(query)) : files;
  const parentPath = dirname(path);
  // Only highlight the selection when it lives in the folder being shown.
  const selectedName = selected && dirname(selected) === path ? selected.slice(selected.lastIndexOf("/") + 1) : null;

  const renderBody = () => {
    if (placeholder) return placeholder;
    if (loading && files.length === 0) return <EmptyState icon={<Spinner className="h-6 w-6" />} title="Loading…" />;
    if (files.length === 0)
      return <EmptyState icon={<FolderIcon className="h-10 w-10" />} title="This folder is empty" />;
    if (visible.length === 0) return <EmptyState title={`Nothing matches “${filter}”`} />;
    return (
      <FileList
        files={visible}
        parentPath={path ? parentPath : null}
        selected={selectedName}
        onOpenFolder={(name) => open(name === ".." ? parentPath : joinPath(path, name))}
        onSelect={onSelect && ((name) => onSelect(joinPath(path, name)))}
        onOpenFile={onOpenFile && ((name) => onOpenFile(joinPath(path, name)))}
      />
    );
  };

  return (
    <Panel className={`flex min-h-0 flex-col ${className}`}>
      <div className="flex items-center gap-2 border-b border-slate-800 px-3 py-2">
        <Breadcrumbs rootLabel={rootLabel} path={path} onNavigate={open} />
        {!placeholder && files.length > 0 && (
          <label className="relative shrink-0">
            <SearchIcon className="pointer-events-none absolute left-2 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-500" />
            <input
              value={filter}
              onChange={(e) => setFilter(e.target.value)}
              placeholder="Filter"
              aria-label="Filter files"
              className="w-36 rounded-md border border-slate-700 bg-slate-950 py-1 pl-7 pr-2 text-sm text-slate-100 placeholder:text-slate-600 focus:border-sky-500 focus:outline-none"
            />
          </label>
        )}
        <Button
          variant="ghost"
          onClick={refresh}
          disabled={Boolean(placeholder) || loading}
          className="shrink-0 px-2 py-1.5"
          title="Refresh"
        >
          <RefreshIcon className={`h-4 w-4 ${loading ? "animate-spin" : ""}`} />
        </Button>
      </div>
      <div className="min-h-0 flex-1 overflow-auto">{renderBody()}</div>
      {!placeholder && files.length > 0 && (
        <div className="border-t border-slate-800 px-4 py-1.5 text-xs text-slate-500">
          {query ? `${visible.length} of ${files.length}` : files.length} {files.length === 1 ? "item" : "items"}
        </div>
      )}
    </Panel>
  );
}

export default FileBrowser;

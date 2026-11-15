import { formatBytes } from "../lib/format.js";
import { BackIcon, FileIcon, FolderIcon } from "./icons.jsx";

/**
 * Lists the entries of one folder by name.
 * - Folders open with a click (onOpenFolder).
 * - Files are selectable when onSelect is given, and open with double-click / Enter (onOpenFile).
 * - parentPath !== null adds a ".." row that navigates up.
 */
function FileList({ files, parentPath = null, selected, onSelect, onOpenFolder, onOpenFile }) {
  const rows = parentPath !== null ? [{ name: "..", isFolder: true, isParent: true }, ...files] : files;

  return (
    <ul className="divide-y divide-slate-800/70">
      {rows.map((file) => {
        const isSelected = !file.isFolder && selected === file.name;
        const clickable = file.isFolder ? Boolean(onOpenFolder) : Boolean(onSelect);
        const activate = () => (file.isFolder ? onOpenFolder?.(file.name) : onOpenFile?.(file.name));

        return (
          <li
            key={file.name}
            tabIndex={clickable ? 0 : undefined}
            aria-selected={onSelect && !file.isFolder ? isSelected : undefined}
            onClick={() => (file.isFolder ? onOpenFolder?.(file.name) : onSelect?.(file.name))}
            onDoubleClick={() => !file.isFolder && activate()}
            onKeyDown={(e) => e.key === "Enter" && activate()}
            className={`flex items-center gap-3 px-4 py-2.5 text-sm outline-none transition
              ${clickable ? "cursor-pointer hover:bg-slate-800/60 focus-visible:bg-slate-800/60" : ""}
              ${isSelected ? "bg-sky-500/15 text-white hover:bg-sky-500/20" : "text-slate-200"}`}
          >
            {file.isParent ? (
              <BackIcon className="h-5 w-5 shrink-0 text-slate-500" />
            ) : file.isFolder ? (
              <FolderIcon className="h-5 w-5 shrink-0 text-amber-400" />
            ) : (
              <FileIcon className={`h-5 w-5 shrink-0 ${isSelected ? "text-sky-300" : "text-slate-400"}`} />
            )}
            <span className="min-w-0 flex-1 truncate" title={file.name}>
              {file.isParent ? "Up one level" : file.name}
            </span>
            {!file.isParent && (
              <span className="shrink-0 tabular-nums text-slate-500">
                {file.isFolder ? "Folder" : formatBytes(file.size)}
              </span>
            )}
          </li>
        );
      })}
    </ul>
  );
}

export default FileList;

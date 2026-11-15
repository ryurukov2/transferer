import { useCallback, useRef, useState } from "react";
import { useToast } from "../components/Toast.jsx";
import { errorMessage } from "./format.js";

/** Joins slash-separated relative paths, treating "" as the root. */
export const joinPath = (dir, name) => (dir ? `${dir}/${name}` : name);

/** Parent of a slash-separated relative path ("" for top-level entries). */
export const dirname = (path) => (path.includes("/") ? path.slice(0, path.lastIndexOf("/")) : "");

/** Last segment of a slash-separated relative path. */
export const basename = (path) => path.slice(path.lastIndexOf("/") + 1);

/**
 * Navigation state for a folder tree served by listFn(relativePath).
 * Returns { path, files, loading, open(path), refresh(), reset() }.
 */
export function useFileBrowser(listFn) {
  const notify = useToast();
  const [path, setPath] = useState("");
  const [files, setFiles] = useState([]);
  const [loading, setLoading] = useState(false);
  // Ignore responses to requests that were superseded by a newer navigation.
  const requestId = useRef(0);

  const open = useCallback(
    async (target = "") => {
      const id = ++requestId.current;
      setLoading(true);
      try {
        const list = (await listFn(target)) ?? [];
        if (id !== requestId.current) return;
        setPath(target);
        setFiles(list);
      } catch (err) {
        if (id === requestId.current) notify(`Couldn't open folder: ${errorMessage(err)}`, { type: "error" });
      } finally {
        if (id === requestId.current) setLoading(false);
      }
    },
    [listFn, notify]
  );

  const refresh = useCallback(() => open(path), [open, path]);

  const reset = useCallback(() => {
    requestId.current++;
    setPath("");
    setFiles([]);
    setLoading(false);
  }, []);

  return { path, files, loading, open, refresh, reset };
}

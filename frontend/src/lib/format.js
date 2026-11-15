export function formatBytes(bytes) {
  if (!bytes) return "0 B";
  const units = ["B", "KB", "MB", "GB", "TB"];
  const i = Math.min(Math.floor(Math.log(bytes) / Math.log(1024)), units.length - 1);
  const value = bytes / 1024 ** i;
  return `${value.toFixed(i === 0 || value >= 100 ? 0 : 1)} ${units[i]}`;
}

/** Wails rejects bound-method promises with a plain string; normalise to a message. */
export function errorMessage(err) {
  return typeof err === "string" ? err : err?.message ?? "Something went wrong";
}

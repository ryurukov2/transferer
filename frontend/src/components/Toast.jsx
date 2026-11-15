import { createContext, useCallback, useContext, useState } from "react";
import { AlertIcon, CheckIcon } from "./icons.jsx";

const ToastContext = createContext(() => {});

/** Returns notify(message, { type: "success" | "error" }). */
export const useToast = () => useContext(ToastContext);

let nextId = 0;

export function ToastProvider({ children }) {
  const [toasts, setToasts] = useState([]);

  const notify = useCallback((message, { type = "success", duration = 4000 } = {}) => {
    const id = nextId++;
    setToasts((t) => [...t, { id, message, type }]);
    setTimeout(() => setToasts((t) => t.filter((x) => x.id !== id)), duration);
  }, []);

  return (
    <ToastContext.Provider value={notify}>
      {children}
      <div className="pointer-events-none fixed bottom-4 right-4 z-50 flex w-80 flex-col gap-2">
        {toasts.map((t) => (
          <div
            key={t.id}
            role="status"
            className="toast-in pointer-events-auto flex items-start gap-3 rounded-lg border border-slate-700 bg-slate-800 p-3 text-sm text-slate-100 shadow-xl"
          >
            {t.type === "error" ? (
              <AlertIcon className="h-5 w-5 shrink-0 text-rose-400" />
            ) : (
              <CheckIcon className="h-5 w-5 shrink-0 text-emerald-400" />
            )}
            <span className="break-words">{t.message}</span>
          </div>
        ))}
      </div>
    </ToastContext.Provider>
  );
}

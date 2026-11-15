import { useEffect, useState } from "react";
import {
  GetLocalAddresses,
  GetSharedDir,
  ListSharedFiles,
  OpenSharedDir,
  SelectSharedDir,
} from "../../wailsjs/go/main/App.js";
import { EventsOn } from "../../wailsjs/runtime/runtime.js";
import FileBrowser from "../components/FileBrowser.jsx";
import { DownloadIcon, UserIcon } from "../components/icons.jsx";
import { useToast } from "../components/Toast.jsx";
import { Button, FolderField, PageHeader, Panel } from "../components/ui.jsx";
import { errorMessage, formatBytes } from "../lib/format.js";
import { basename, useFileBrowser } from "../lib/useFileBrowser.js";

const MAX_ACTIVITY = 50;

/** Tracks connected clients and recent downloads from "server-activity" events. */
function useServerActivity() {
  const [clients, setClients] = useState({}); // ip -> open connection count
  const [events, setEvents] = useState([]);

  useEffect(
    () =>
      EventsOn("server-activity", (act) => {
        if (act.kind === "connected" || act.kind === "disconnected") {
          setClients((c) => {
            const n = (c[act.client] ?? 0) + (act.kind === "connected" ? 1 : -1);
            const { [act.client]: _, ...rest } = c;
            return n > 0 ? { ...rest, [act.client]: n } : rest;
          });
        }
        if (act.kind !== "disconnected") {
          setEvents((e) => [{ ...act, at: new Date() }, ...e].slice(0, MAX_ACTIVITY));
        }
      }),
    []
  );

  return { clients: Object.keys(clients), events };
}

function ActivityFeed({ clients, events }) {
  return (
    <Panel className="flex w-72 shrink-0 flex-col">
      <div className="border-b border-slate-800 px-4 py-2.5 text-sm font-medium text-slate-300">
        Activity
        <span className="ml-2 font-normal text-slate-500">
          {clients.length} {clients.length === 1 ? "computer" : "computers"} connected
        </span>
      </div>
      <div className="min-h-0 flex-1 overflow-auto">
        {events.length === 0 ? (
          <div className="px-4 py-8 text-center text-sm text-slate-500">
            Nothing yet. Downloads from other computers will show up here.
          </div>
        ) : (
          <ul className="divide-y divide-slate-800/70">
            {events.map((e, i) => (
              <li key={i} className="flex gap-3 px-4 py-2.5 text-sm">
                {e.kind === "download" ? (
                  <DownloadIcon className="mt-0.5 h-4 w-4 shrink-0 text-sky-400" />
                ) : (
                  <UserIcon className="mt-0.5 h-4 w-4 shrink-0 text-emerald-400" />
                )}
                <div className="min-w-0 flex-1">
                  <div className="truncate text-slate-200" title={e.path}>
                    {e.kind === "download" ? basename(e.path) : "Connected"}
                  </div>
                  <div className="text-xs text-slate-500">
                    <span className="font-mono">{e.client}</span>
                    {e.kind === "download" && ` · ${formatBytes(e.size)}`} · {e.at.toLocaleTimeString()}
                  </div>
                </div>
              </li>
            ))}
          </ul>
        )}
      </div>
    </Panel>
  );
}

function ServerView({ onBack }) {
  const notify = useToast();
  const [addresses, setAddresses] = useState([]);
  const [sharedDir, setSharedDir] = useState("");
  const browser = useFileBrowser(ListSharedFiles);
  const { clients, events } = useServerActivity();

  useEffect(() => {
    GetLocalAddresses().then((a) => setAddresses(a ?? []));
    GetSharedDir().then(setSharedDir);
    browser.open("");
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const changeFolder = async () => {
    try {
      const dir = await SelectSharedDir();
      if (dir) {
        setSharedDir(dir);
        browser.open("");
      }
    } catch (err) {
      notify(errorMessage(err), { type: "error" });
    }
  };

  const openFolder = () => OpenSharedDir().catch((err) => notify(errorMessage(err), { type: "error" }));

  return (
    <div className="flex h-full flex-col">
      <PageHeader
        title="Sharing files"
        onBack={onBack}
        subtitle={
          <span className="flex items-center gap-2">
            <span className="relative flex h-2 w-2">
              <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-emerald-400 opacity-60" />
              <span className="relative inline-flex h-2 w-2 rounded-full bg-emerald-400" />
            </span>
            Visible on the network
            {addresses.length > 0 && <span className="font-mono text-slate-500">· {addresses.join(", ")}</span>}
          </span>
        }
      />

      <div className="flex min-h-0 flex-1 gap-4 p-6">
        <div className="flex min-w-0 flex-1 flex-col gap-4">
          <Panel className="p-4">
            <FolderField label="Shared folder" path={sharedDir} onChange={changeFolder}>
              <Button variant="ghost" onClick={openFolder} className="shrink-0">
                Open
              </Button>
            </FolderField>
          </Panel>
          <FileBrowser rootLabel="Shared folder" browser={browser} className="flex-1" />
        </div>
        <ActivityFeed clients={clients} events={events} />
      </div>
    </div>
  );
}

export default ServerView;

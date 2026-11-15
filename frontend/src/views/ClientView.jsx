import { useCallback, useEffect, useState } from "react";
import {
  CancelDownload,
  Connect,
  DiscoverServers,
  DownloadFile,
  GetRemoteDir,
  GetSaveDir,
  ListRemoteFiles,
  OpenSaveDir,
  SelectSaveDir,
} from "../../wailsjs/go/main/App.js";
import { EventsOn } from "../../wailsjs/runtime/runtime.js";
import FileBrowser from "../components/FileBrowser.jsx";
import { DownloadIcon, RefreshIcon, ServerIcon, XIcon } from "../components/icons.jsx";
import { useToast } from "../components/Toast.jsx";
import { Button, EmptyState, FolderField, PageHeader, Panel, Spinner } from "../components/ui.jsx";
import { errorMessage, formatBytes } from "../lib/format.js";
import { basename, useFileBrowser } from "../lib/useFileBrowser.js";

function ServerSidebar({ servers, scanning, connected, connecting, onScan, onConnect }) {
  const [manualAddr, setManualAddr] = useState("");

  const submitManual = (e) => {
    e.preventDefault();
    const addr = manualAddr.trim();
    if (addr) onConnect(addr.includes(":") ? addr : `${addr}:8888`);
  };

  return (
    <Panel className="flex w-64 shrink-0 flex-col">
      <div className="flex items-center justify-between border-b border-slate-800 px-4 py-2.5">
        <span className="text-sm font-medium text-slate-300">Computers</span>
        <Button variant="ghost" onClick={onScan} disabled={scanning} className="px-2 py-1.5" title="Scan again">
          <RefreshIcon className={`h-4 w-4 ${scanning ? "animate-spin" : ""}`} />
        </Button>
      </div>

      <div className="min-h-0 flex-1 overflow-auto p-2">
        {servers.length === 0 ? (
          <div className="px-2 py-6 text-center text-sm text-slate-500">
            {scanning ? "Searching the network…" : "No computers found. Start sharing on the other computer, then scan again."}
          </div>
        ) : (
          servers.map((addr) => {
            const active = addr === connected || addr === connecting;
            return (
              <button
                key={addr}
                type="button"
                onClick={() => onConnect(addr)}
                className={`flex w-full items-center gap-3 rounded-lg px-3 py-2.5 text-left text-sm transition
                  ${active ? "bg-sky-500/15 text-white" : "text-slate-300 hover:bg-slate-800"}`}
              >
                <ServerIcon className={`h-5 w-5 shrink-0 ${active ? "text-sky-300" : "text-slate-500"}`} />
                <span className="flex-1 truncate font-mono">{addr}</span>
                {addr === connecting && <Spinner className="h-3.5 w-3.5" />}
                {addr === connected && <span className="h-2 w-2 rounded-full bg-emerald-400" title="Connected" />}
              </button>
            );
          })
        )}
      </div>

      <form onSubmit={submitManual} className="border-t border-slate-800 p-3">
        <label className="text-xs font-medium uppercase tracking-wide text-slate-500" htmlFor="manual-addr">
          Connect by address
        </label>
        <div className="mt-1.5 flex gap-2">
          <input
            id="manual-addr"
            value={manualAddr}
            onChange={(e) => setManualAddr(e.target.value)}
            placeholder="192.168.1.20"
            className="min-w-0 flex-1 rounded-lg border border-slate-700 bg-slate-950 px-2.5 py-1.5 font-mono text-sm text-slate-100 placeholder:text-slate-600 focus:border-sky-500 focus:outline-none"
          />
          <Button type="submit" className="px-3 py-1.5">
            Go
          </Button>
        </div>
      </form>
    </Panel>
  );
}

function DownloadProgress({ path, done, total, onCancel }) {
  const pct = total ? Math.round((done / total) * 100) : 0;
  return (
    <div>
      <div className="mb-1.5 flex items-center gap-2 text-sm text-slate-300">
        <span className="min-w-0 flex-1 truncate">
          Downloading {basename(path)}… <span className="text-slate-500">{pct}%</span>
        </span>
        <Button variant="ghost" onClick={onCancel} className="shrink-0 px-2 py-1 text-xs">
          <XIcon className="h-3.5 w-3.5" /> Cancel
        </Button>
      </div>
      <div className="flex items-center gap-3 text-xs text-slate-400">
      <div className="h-1.5 flex-1 overflow-hidden rounded-full bg-slate-800">
        <div className="h-full rounded-full bg-sky-500 transition-[width]" style={{ width: `${pct}%` }} />
      </div>
      <span className="w-36 text-right tabular-nums">
        {formatBytes(done)} / {formatBytes(total)}
        </span>
      </div>
    </div>
  );
}

function ClientView({ onBack }) {
  const notify = useToast();
  const [servers, setServers] = useState([]);
  const [scanning, setScanning] = useState(false);
  const [connected, setConnected] = useState(null);
  const [connecting, setConnecting] = useState(null);
  const [remoteDir, setRemoteDir] = useState("");
  const [selected, setSelected] = useState(null); // relative path of the selected file
  const [saveDir, setSaveDir] = useState("");
  const [progress, setProgress] = useState(null);
  const browser = useFileBrowser(ListRemoteFiles);
  const { open: openFolder, reset: resetBrowser } = browser;

  const connect = useCallback(
    async (addr, folder = "") => {
      setConnecting(addr);
      setSelected(null);
      try {
        await Connect(addr);
        setConnected(addr);
        setRemoteDir(await GetRemoteDir());
        await openFolder(folder);
      } catch (err) {
        setConnected(null);
        resetBrowser();
        notify(errorMessage(err), { type: "error" });
      } finally {
        setConnecting(null);
      }
    },
    [openFolder, resetBrowser, notify]
  );

  const scan = useCallback(async () => {
    setScanning(true);
    try {
      const found = (await DiscoverServers()) ?? [];
      setServers(found);
      if (found.length === 1) connect(found[0]);
    } catch (err) {
      notify(`Network scan failed: ${errorMessage(err)}`, { type: "error" });
    } finally {
      setScanning(false);
    }
  }, [connect, notify]);

  useEffect(() => {
    GetSaveDir().then(setSaveDir);
    scan();
    return EventsOn("download-progress", setProgress);
    // Run once on mount; scan is intentionally not re-triggered.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const download = async (path = selected) => {
    if (!path || progress) return;
    setProgress({ path, done: 0, total: 0 });
    try {
      const saved = await DownloadFile(path);
      notify(`Saved ${saved}`);
    } catch (err) {
      const msg = errorMessage(err);
      const cancelled = msg === "download cancelled";
      notify(cancelled ? "Download cancelled" : `Download failed: ${msg}`, { type: cancelled ? "success" : "error" });
      // An interrupted transfer drops the connection; reconnect where we were.
      if (connected) connect(connected, browser.path);
    } finally {
      setProgress(null);
    }
  };

  const changeSaveDir = async () => {
    try {
      const dir = await SelectSaveDir();
      if (dir) setSaveDir(dir);
    } catch (err) {
      notify(errorMessage(err), { type: "error" });
    }
  };

  const openSaveDir = () => OpenSaveDir().catch((err) => notify(errorMessage(err), { type: "error" }));

  const placeholder = !connected && (
    <EmptyState icon={<ServerIcon className="h-10 w-10" />} title={connecting ? "Connecting…" : "Not connected"}>
      {!connecting && "Pick a computer from the list on the left to see what it's sharing."}
    </EmptyState>
  );

  return (
    <div className="flex h-full flex-col">
      <PageHeader
        title="Receiving files"
        subtitle={connected ? `Connected to ${connected} · ${remoteDir}` : "Not connected"}
        onBack={onBack}
      />

      <div className="flex min-h-0 flex-1 gap-4 p-6">
        <ServerSidebar
          servers={servers}
          scanning={scanning}
          connected={connected}
          connecting={connecting}
          onScan={scan}
          onConnect={connect}
        />

        <div className="flex min-w-0 flex-1 flex-col gap-4">
          <FileBrowser
            rootLabel={connected ? connected.split(":")[0] : "Server"}
            browser={browser}
            selected={selected}
            onSelect={setSelected}
            onOpenFile={download}
            placeholder={placeholder}
            className="flex-1"
          />

          <Panel className="flex flex-col gap-3 p-4">
            <FolderField label="Save to" path={saveDir} onChange={changeSaveDir}>
              <Button variant="ghost" onClick={openSaveDir} className="shrink-0">
                Open
              </Button>
            </FolderField>
            {progress ? (
              <DownloadProgress {...progress} onCancel={CancelDownload} />
            ) : (
              <Button variant="primary" onClick={() => download()} disabled={!selected} className="w-full py-2.5">
                <DownloadIcon className="h-4 w-4" />
                <span className="truncate">
                  {selected ? `Download ${basename(selected)}` : "Select a file to download"}
                </span>
              </Button>
            )}
            <p className="text-center text-xs text-slate-500">
              Click a folder to open it · double-click a file to download it
            </p>
          </Panel>
        </div>
      </div>
    </div>
  );
}

export default ClientView;

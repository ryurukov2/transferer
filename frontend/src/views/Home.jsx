import { DownloadIcon, ShareIcon } from "../components/icons.jsx";
import { Spinner } from "../components/ui.jsx";

function ModeCard({ icon, title, description, accent, onClick, busy }) {
  return (
    <button
      type="button"
      onClick={onClick}
      disabled={busy}
      className="group flex flex-col items-start gap-4 rounded-2xl border border-slate-800 bg-slate-900/60 p-6 text-left transition
        hover:-translate-y-0.5 hover:border-slate-600 hover:bg-slate-900 focus:outline-none focus-visible:ring-2 focus-visible:ring-sky-400
        disabled:cursor-wait"
    >
      <div className={`rounded-xl p-3 ${accent}`}>{busy ? <Spinner className="h-6 w-6" /> : icon}</div>
      <div>
        <div className="text-lg font-semibold text-white">{title}</div>
        <div className="mt-1 text-sm text-slate-400">{description}</div>
      </div>
    </button>
  );
}

function Home({ onShare, onReceive, starting }) {
  return (
    <div className="flex h-full flex-col items-center justify-center px-6">
      <h1 className="text-3xl font-bold tracking-tight text-white">Transferer</h1>
      <p className="mt-2 text-slate-400">Send files between computers on your local network.</p>

      <div className="mt-10 grid w-full max-w-2xl grid-cols-1 gap-4 sm:grid-cols-2">
        <ModeCard
          icon={<ShareIcon className="h-6 w-6" />}
          title="Share files"
          description="Make a folder on this computer available to others on the network."
          accent="bg-sky-500/15 text-sky-300"
          onClick={onShare}
          busy={starting}
        />
        <ModeCard
          icon={<DownloadIcon className="h-6 w-6" />}
          title="Receive files"
          description="Find a computer that's sharing and download files from it."
          accent="bg-emerald-500/15 text-emerald-300"
          onClick={onReceive}
        />
      </div>
    </div>
  );
}

export default Home;

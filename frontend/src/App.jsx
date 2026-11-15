import { useState } from "react";
import { StartServer, Stop } from "../wailsjs/go/main/App.js";
import { ToastProvider, useToast } from "./components/Toast.jsx";
import { errorMessage } from "./lib/format.js";
import ClientView from "./views/ClientView.jsx";
import Home from "./views/Home.jsx";
import ServerView from "./views/ServerView.jsx";

function Screens() {
  const notify = useToast();
  const [mode, setMode] = useState(null); // null | "server" | "client"
  const [starting, setStarting] = useState(false);

  const share = async () => {
    setStarting(true);
    try {
      await StartServer();
      setMode("server");
    } catch (err) {
      notify(`Couldn't start sharing: ${errorMessage(err)}`, { type: "error", duration: 6000 });
    } finally {
      setStarting(false);
    }
  };

  const goHome = async () => {
    await Stop();
    setMode(null);
  };

  if (mode === "server") return <ServerView onBack={goHome} />;
  if (mode === "client") return <ClientView onBack={goHome} />;
  return <Home onShare={share} onReceive={() => setMode("client")} starting={starting} />;
}

function App() {
  return (
    <ToastProvider>
      <Screens />
    </ToastProvider>
  );
}

export default App;

import { useState, useEffect } from "react";
import logo from "./assets/images/logo-universal.png";
import "./App.css";
import ServerLayout from "./Server.jsx";
import ClientLayout from "./Client.jsx";
import { EventsEmit } from "/wailsjs/runtime/runtime.js";
// import { reqFile } from "../wailsjs/go/main/App"
function App() {
  const [role, setRole] = useState(null);

  const handleServerClick = () => {
    setRole("server");
    EventsEmit("start-server");
    console.log("serverstart");
  };
  
  const handleClientClick = () => {
    setRole("client");
    EventsEmit("start-client");
    console.log("clientstart");
  };

  useEffect(() => {
    const handleBeforeUnload = () => {
      EventsEmit('stop-servers');
    };

    window.addEventListener('beforeunload', handleBeforeUnload);

    return () => {
      window.removeEventListener('beforeunload', handleBeforeUnload);
    };
  }, []);

  return (
    <div className="h-screen flex flex-col justify-center items-center">
      {!role && (
        <div className="text-center">
          <h1 className="text-4xl mb-6">File Transfer App</h1>
          <div className="space-x-4">
            <button
              onClick={handleServerClick}
              className="bg-blue-500 text-white py-2 px-4 rounded"
            >
              Run as Server
            </button>
            <button
              onClick={handleClientClick}
              className="bg-green-500 text-white py-2 px-4 rounded"
            >
              Run as Client
            </button>
          </div>
        </div>
      )}

      {role === "server" && (
        <div className="text-center">
          <h2 className="text-2xl">Server Mode</h2>
            <ServerLayout/>
        </div>
      )}

      {role === "client" && (
        <div className="text-center">
          <h2 className="text-2xl">Client Mode</h2>
          <ClientLayout/>
        </div>
      )}
    </div>
  );
}
export default App;

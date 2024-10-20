import { useState, useEffect } from "react";
import ServerList from "./ServerList.jsx";
import FileList from "./FileList.jsx";
import HomeButton from "./HomeButton.jsx";
import {
  ReqFile,
  DiscServers,
  GetFilesFromClient,
  SetClientConnection,
  GetServerDirectory,
} from "../wailsjs/go/main/App.js";
function ClientLayout() {
  const [servers, setServers] = useState([]);
  const [fileDirectory, setFileDirectory] = useState("")
  const [serverAvailable, setServerAvailable] = useState(false);
  const [selectedServer, setSelectedServer] = useState(null);
  const [files, setFiles] = useState([]);
  const [connAvailable, setConnAvailable] = useState(false);
  const [result, setResult] = useState("");
  const [selectedFile, setSelectedFile] = useState(null);
  const reqButtonDisableStatus = selectedFile == null;
  const updateResult = (res) => setResult(res);
  const handleRequestFile = () => {
    if (selectedFile) {
      ReqFile(selectedFile.name).then(updateResult);
    }
  };
  const serverDirScan = async () => {
    if (selectedServer != null){
      const fileDir = await GetServerDirectory();
      console.log(fileDir);
      setFileDirectory(fileDir);
    }
  }
  const fileScan = async () => {
    if (selectedServer != null) {
      const availableFiles = await GetFilesFromClient();
      console.log(availableFiles);
      if (availableFiles.length != 0) {
        setFiles(availableFiles);
      }
    }
  };
  const serverScan = async () => {
    const availableServers = await DiscServers();
    const numberOfServers = availableServers.length;
    if (numberOfServers != 0) {
      setServers(availableServers);
      setServerAvailable(true);

      if (numberOfServers == 1) {
        setSelectedServer(availableServers[0]);
      }
    }
  };
  useEffect(() => {
    serverScan();
  }, []);
  useEffect(() => {
    if(selectedServer){
      SetClientConnection(selectedServer).then(setConnAvailable);
    }
  }, [selectedServer]);
  useEffect(() => {
    serverDirScan()
    fileScan()
  }, [connAvailable])
  return (
    <div className="flex flex-col h-full px-8 w-full">
      <div className="items-center justify-between flex flex-row">
        <HomeButton/>
        <h2 className="text-2xl p-4">Client Mode</h2>
        <div></div>
      </div>
      {!serverAvailable ? (
        <div className="flex flex-col items-start">
          <div className="text-gray-600 mb-4">
            No servers available. Make sure the server is launched on the host
            and try again.
          </div>
          <button
            type="button"
            onClick={serverScan}
            className="bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600"
          >
            Scan
          </button>
        </div>
      ) : (
        <div className="flex flex-row gap-4 h-5/6">
          <div className="w-1/4">
            <div className="bg-blue-800 p-4 rounded-lg overflow-clip">
              <ServerList servers={servers}
              selectedServer={selectedServer}
              setSelectedServer={setSelectedServer} />
            </div>
          </div>

          <div className="w-3/4 flex flex-col h-full">
            <div className="bg-grey-400 p-4 rounded-lg flex-1 overflow-auto">
              <div><strong>File Directory: </strong>{fileDirectory}</div>
              <FileList
                files={files}
                selectedFile={selectedFile}
                setSelectedFile={setSelectedFile}
              />
            </div>
            <div className="">
            {selectedFile && (
            <div>
              <strong>Selected File:</strong> {selectedFile.name}
            </div>
          )}
            <button
              type="submit"
              id="reqFileSubmit"
              onClick={handleRequestFile}
              disabled={reqButtonDisableStatus}
              className="bg-green-500 text-white px-4 py-2 rounded hover:bg-green-600 w-full"
            >
              Request file
            </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

export default ClientLayout;

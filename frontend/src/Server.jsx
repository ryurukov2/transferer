import HomeButton from "./HomeButton";
import FileList from "./FileList.jsx";
import {
  GetFilesFromServer,
  SelectServerFolder,
  GetCurrentServerDir,
} from "../wailsjs/go/main/App.js";
import { useEffect, useRef, useState } from "react";
function ServerLayout() {
  const [files, setFiles] = useState([]);
  const [selectedFile, setSelectedFile] = useState(null);
  const [folderPath, setFolderPath] = useState(".");
  const getCurrentServerDir = async () => {
    const currentDir = await GetCurrentServerDir();
    setFolderPath(currentDir);
  };
  const getFilesFromServ = async () => {
    const availableFiles = await GetFilesFromServer(folderPath);
    console.log(availableFiles);
    if (availableFiles.length != 0) {
      setFiles(availableFiles);
    }
  };

  const handleFolderSelect = async () => {
    try {
      const path = await SelectServerFolder();
      setFolderPath(path);
      console.log("Selected Folder Path:", path);
    } catch (error) {
      console.error("Failed to select folder:", error);
    }
  };

  useEffect(() => {
    getFilesFromServ();
  }, [folderPath]);
  useEffect(() => {
    getCurrentServerDir();
  }, []);

  return (
    <div className="flex flex-col w-full h-full px-8">
      <div className="items-center justify-between flex flex-row py-4">
        <HomeButton />
        <h2 className="text-2xl">Server Mode</h2>
        <button onClick={handleFolderSelect} className="btn btn-primary">
          Select Folder
        </button>
      </div>
      <div className="flex flex-row gap-4 h-5/6 justify-center">
        <div className="w-3/4 flex flex-col h-full">
          <div>{folderPath && <p>Selected Folder: {folderPath}</p>}</div>
          <div className="bg-grey-400 p-4 rounded-lg flex-1 overflow-auto">
            <FileList files={files} />
          </div>
        </div>
      </div>
    </div>
  );
}

export default ServerLayout;

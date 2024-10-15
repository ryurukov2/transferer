import { useState } from "react";
import fileImage from "./assets/images/file.png"
import folderImage from "./assets/images/folder.png"
function FileList({ files, selectedFile, setSelectedFile }) {

  
    const handleFileClick = (file) => {
      setSelectedFile(file);
    };
  
    return (
        <>
          Files:
          {files.map((file, index) => (
            <div
              key={index}
              onClick={() => handleFileClick(file)}
              style={{
                cursor: 'pointer',
                backgroundColor: selectedFile === file ? 'lightblue' : 'transparent',
                border: selectedFile === file ? '1px solid blue' : '1px solid transparent',
              }}
              className="flex items-center"
            >
              {/* Display the appropriate image depending on whether it's a file or folder */}
              <img
                src={file.isFolder ? folderImage : fileImage}
                alt={file.isFolder ? 'Folder' : 'File'}
                style={{ width: '24px', height: '24px', marginRight: '10px' }}
              />
              {file.name}
            </div>
          ))}
        </>
      );
  }

export default FileList
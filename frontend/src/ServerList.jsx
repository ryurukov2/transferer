function ServerList({ servers, selectedServer, setSelectedServer }) {
  return (
    <>
      {servers.map((serverIP, index) => (
        <div key={index}
        onClick={() => setSelectedServer(serverIP)}
        className="cursor-pointer"
        style={{
          backgroundColor: selectedServer===serverIP ? 'darkblue' : 'transparent'
        }}>{serverIP}</div>
      ))}
    </>
  );
}

export default ServerList;

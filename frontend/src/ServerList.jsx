function ServerList({ servers }) {
  return (
    <>
      {servers.map((serverIP, index) => (
        <div key={index}>{serverIP}</div>
      ))}
    </>
  );
}

export default ServerList;

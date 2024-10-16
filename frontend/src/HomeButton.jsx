function HomeButton() {
  return (
    <>
      <button className="cursor-pointer" onClick={() => {
        window.location.reload()
      }}>Home</button>
    </>
  );
}

export default HomeButton;

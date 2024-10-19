function HomeButton() {
  return (
    <>
      <button className="cursor-pointer btn btn-primary" onClick={() => {
        window.location.reload()
      }}>Home</button>
    </>
  );
}

export default HomeButton;

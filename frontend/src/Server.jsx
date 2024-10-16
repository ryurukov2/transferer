import HomeButton from "./HomeButton";

function ServerLayout(){
    return <div className="flex flex-col w-full h-full px-8">
    <div className="items-center justify-between flex flex-row">
        <HomeButton/>
        <h2 className="text-2xl">Server Mode</h2>
        <div></div>
    </div>
    </div>
}

export default ServerLayout;
import ChatWithTextarea from "./ChatWithTextarea";
import "reachat/styles.css"; // glob-import Reachat defaults
import "./index.css"; // Tailwind

function App() {
  return (
    <main className="max-w-3xl mx-auto py-10 px-4"> {/* Added padding */}
      <h1 className="text-2xl font-bold mb-4">Blog post draft</h1>
      <ChatWithTextarea />
      {/* Add more instances or other content here */}
      {/* <h2 className="text-xl font-semibold mt-8 mb-4">Another Text Area</h2> */}
      {/* <ChatWithTextarea /> */}
    </main>
  );
}

export default App; 
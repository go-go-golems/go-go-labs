import "./index.css";
import { Composition } from "remotion";
import { FlowScene } from "./FlowScene";
import { TwitterScraperScene } from "./TwitterScraperScene";

// Each <Composition> is an entry in the sidebar!

export const RemotionRoot: React.FC = () => {
  return (
    <>
      <Composition
        // You can take the "id" to render a video:
        // npx remotion render src/index.ts <id> out/video.mp4
        id="flow"
        component={FlowScene}
        durationInFrames={600}
        fps={30}
        width={1920}
        height={1080}
      />
      <Composition
        id="twitter-scraper"
        component={TwitterScraperScene}
        durationInFrames={150}
        fps={30}
        width={1920}
        height={1080}
      />
    </>
  );
};

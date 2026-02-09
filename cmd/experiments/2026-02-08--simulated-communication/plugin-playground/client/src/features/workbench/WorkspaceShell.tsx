import { Button } from "@/components/ui/button";

interface WorkspaceShellProps {
  customCode: string;
  error: string;
  onCustomCodeChange: (value: string) => void;
  onLoadCustom: () => void;
}

export function WorkspaceShell({
  customCode,
  error,
  onCustomCodeChange,
  onLoadCustom,
}: WorkspaceShellProps) {
  return (
    <div className="border border-cyan-400/30 rounded-sm p-4 bg-card/50 h-full min-h-0 flex flex-col">
      <h2 className="text-lg font-bold text-cyan-400 mb-4 font-mono">WORKSPACE</h2>
      <textarea
        value={customCode}
        onChange={(e) => onCustomCodeChange(e.target.value)}
        placeholder="definePlugin(({ ui }) => { ... })"
        className="w-full flex-1 min-h-[12rem] bg-background/50 border border-cyan-400/20 rounded p-2 font-mono text-xs text-foreground resize-none focus:outline-none focus:border-cyan-400"
      />
      <Button onClick={onLoadCustom} className="w-full mt-2 font-mono text-xs">
        LOAD PLUGIN
      </Button>
      {error && <div className="mt-2 text-red-400 text-xs font-mono">{error}</div>}
    </div>
  );
}

import { Button } from "@/components/ui/button";
import type { PluginDefinition } from "@/lib/presetPlugins";
import type { LoadedPluginMap } from "./types";

interface CatalogShellProps {
  presets: PluginDefinition[];
  loadedCountsByPackage: Record<string, number>;
  loadedPlugins: string[];
  pluginMetaById: LoadedPluginMap;
  onLoadPreset: (presetId: string) => void;
  onUnloadPlugin: (instanceId: string) => void;
}

export function CatalogShell({
  presets,
  loadedCountsByPackage,
  loadedPlugins,
  pluginMetaById,
  onLoadPreset,
  onUnloadPlugin,
}: CatalogShellProps) {
  return (
    <div className="border border-cyan-400/30 rounded-sm p-4 bg-card/50 h-full min-h-0 flex flex-col">
      <h2 className="text-lg font-bold text-cyan-400 mb-4 font-mono">CATALOG</h2>
      <div className="flex-1 min-h-0 overflow-y-auto pr-1">
        <div className="space-y-2">
          {presets.map((preset) => (
            <Button
              key={preset.id}
              onClick={() => onLoadPreset(preset.id)}
              variant={(loadedCountsByPackage[preset.id] ?? 0) > 0 ? "default" : "outline"}
              className="w-full justify-start font-mono text-xs"
            >
              {preset.title}
              {(loadedCountsByPackage[preset.id] ?? 0) > 0 &&
                ` (${loadedCountsByPackage[preset.id]})`}
            </Button>
          ))}
        </div>

        <div className="mt-6 border-t border-cyan-400/20 pt-4">
          <h3 className="text-sm font-bold text-cyan-400 mb-2 font-mono">LOADED</h3>
          <div className="space-y-1">
            {loadedPlugins.map((instanceId) => (
              <div key={instanceId} className="flex items-center justify-between text-xs font-mono">
                <span>
                  {pluginMetaById[instanceId]?.title ?? "Plugin"} [{instanceId}]
                </span>
                <button
                  onClick={() => onUnloadPlugin(instanceId)}
                  className="text-red-400 hover:text-red-300"
                >
                  X
                </button>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}

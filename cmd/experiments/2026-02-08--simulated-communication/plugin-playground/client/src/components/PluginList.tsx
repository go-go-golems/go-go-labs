// Design Philosophy: Technical Brutalism - Vertical plugin list with status indicators
// Glowing borders for active plugins, monospace labels

import React from "react";
import { useSelector, useDispatch } from "react-redux";
import type { RootState } from "@/store/store";
import { pluginToggled, pluginRemoved } from "@/store/store";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Power, Trash2, Code } from "lucide-react";

interface PluginListProps {
  onSelectPlugin: (pluginId: string) => void;
  selectedPluginId: string | null;
}

export function PluginList({ onSelectPlugin, selectedPluginId }: PluginListProps) {
  const plugins = useSelector((state: RootState) => state.plugins.plugins);
  const dispatch = useDispatch();

  const pluginList = Object.values(plugins);

  return (
    <div className="flex flex-col h-full border-r border-accent/30 bg-card/30">
      <div className="px-4 py-3 border-b border-accent/30 bg-accent/5">
        <h2 className="font-mono text-sm uppercase tracking-wider font-bold text-accent">
          Loaded Plugins
        </h2>
        <p className="font-mono text-xs text-muted-foreground mt-1">
          {pluginList.length} active
        </p>
      </div>

      <div className="flex-1 overflow-y-auto p-2">
        {pluginList.length === 0 ? (
          <div className="p-4 text-center">
            <p className="font-mono text-xs text-muted-foreground">
              No plugins loaded
            </p>
            <p className="font-mono text-xs text-muted-foreground mt-2">
              Load a preset to get started
            </p>
          </div>
        ) : (
          <div className="space-y-2">
            {pluginList.map((plugin) => (
              <div
                key={plugin.id}
                className={`
                  border rounded-sm p-3 transition-all cursor-pointer
                  ${
                    selectedPluginId === plugin.id
                      ? "border-accent bg-accent/10 shadow-[0_0_15px_rgba(0,255,255,0.3)]"
                      : "border-accent/30 bg-card/50 hover:border-accent/50"
                  }
                `}
                onClick={() => onSelectPlugin(plugin.id)}
              >
                <div className="flex items-start justify-between mb-2">
                  <div className="flex-1 min-w-0">
                    <div className="font-mono text-sm font-bold text-foreground truncate">
                      {plugin.meta.title || plugin.id}
                    </div>
                    <div className="font-mono text-xs text-muted-foreground mt-1">
                      {plugin.meta.widgets.length} widget(s)
                    </div>
                  </div>
                  
                  <Badge
                    variant={plugin.status === "loaded" ? "default" : "outline"}
                    className={`
                      font-mono text-xs uppercase tracking-wider ml-2
                      ${plugin.status === "loaded" ? "bg-accent/20 text-accent border-accent/50" : ""}
                      ${plugin.status === "error" ? "bg-destructive/20 text-destructive border-destructive/50" : ""}
                      ${plugin.status === "loading" ? "bg-yellow-500/20 text-yellow-500 border-yellow-500/50" : ""}
                    `}
                  >
                    {plugin.status}
                  </Badge>
                </div>

                {plugin.error && (
                  <div className="font-mono text-xs text-destructive mb-2 truncate">
                    Error: {plugin.error}
                  </div>
                )}

                <div className="flex items-center gap-2">
                  <Button
                    onClick={(e) => {
                      e.stopPropagation();
                      dispatch(pluginToggled(plugin.id));
                    }}
                    size="sm"
                    variant="ghost"
                    className="font-mono text-xs h-7 px-2"
                  >
                    <Power className={`w-3 h-3 ${plugin.enabled ? "text-accent" : "text-muted-foreground"}`} />
                  </Button>
                  
                  <Button
                    onClick={(e) => {
                      e.stopPropagation();
                      if (confirm(`Remove plugin "${plugin.id}"?`)) {
                        dispatch(pluginRemoved(plugin.id));
                      }
                    }}
                    size="sm"
                    variant="ghost"
                    className="font-mono text-xs h-7 px-2 text-destructive hover:text-destructive"
                  >
                    <Trash2 className="w-3 h-3" />
                  </Button>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

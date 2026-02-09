/**
 * Sidebar wired to the RTK store.
 *
 * Reads workbench + runtime state, dispatches workbench actions.
 * The underlying <Sidebar> is a pure presentational component that can
 * be used in Storybook without a store.
 */
import { useAppDispatch, useAppSelector, toggleSidebar, focusInstance, selectLoadedPluginIds } from "@/store";
import { Sidebar, type CatalogEntry, type RunningInstance } from "./Sidebar";
import { presetPlugins } from "@/lib/presetPlugins";
import type { RootState } from "@/store";

// ---------------------------------------------------------------------------
// Build catalog entries from preset definitions
// ---------------------------------------------------------------------------

function capSummary(preset: (typeof presetPlugins)[number]): string {
  const r = (preset.capabilities?.readShared?.length ?? 0) > 0;
  const w = (preset.capabilities?.writeShared?.length ?? 0) > 0;
  if (r && w) return "R/W";
  if (r) return "R";
  return "";
}

const CATALOG: CatalogEntry[] = presetPlugins.map((p) => ({
  id: p.id,
  title: p.title,
  description: p.description,
  capabilitySummary: capSummary(p),
}));

// ---------------------------------------------------------------------------
// Selector: build RunningInstance[] from runtime state
// ---------------------------------------------------------------------------

function selectRunningInstances(state: RootState): RunningInstance[] {
  const ids = selectLoadedPluginIds(state);
  const plugins = state.runtime.plugins;
  const grants = state.runtime.grantsByInstance;

  return ids.map((id) => {
    const plugin = plugins[id];
    const grant = grants[id];
    return {
      instanceId: id,
      title: plugin?.title ?? "Plugin",
      packageId: plugin?.packageId ?? "unknown",
      shortId: id.length > 10 ? id.slice(0, 10) : id,
      status: (plugin?.status ?? "loaded") as "loaded" | "error",
      readGrants: grant?.readShared ?? [],
      writeGrants: grant?.writeShared ?? [],
    };
  });
}

// ---------------------------------------------------------------------------
// Connected component
// ---------------------------------------------------------------------------

export interface ConnectedSidebarProps {
  onLoadPreset?: (presetId: string) => void;
  onUnloadInstance?: (instanceId: string) => void;
  onNewPlugin?: () => void;
}

export function ConnectedSidebar({
  onLoadPreset,
  onUnloadInstance,
  onNewPlugin,
}: ConnectedSidebarProps) {
  const dispatch = useAppDispatch();
  const collapsed = useAppSelector((s) => s.workbench.sidebarCollapsed);
  const focusedId = useAppSelector((s) => s.workbench.focusedInstanceId);
  const running = useAppSelector(selectRunningInstances);

  return (
    <Sidebar
      catalog={CATALOG}
      running={running}
      collapsed={collapsed}
      focusedInstanceId={focusedId ?? undefined}
      onToggleCollapse={() => dispatch(toggleSidebar())}
      onFocusInstance={(id) => dispatch(focusInstance(id))}
      onLoadPreset={onLoadPreset}
      onUnloadInstance={onUnloadInstance}
      onNewPlugin={onNewPlugin}
    />
  );
}

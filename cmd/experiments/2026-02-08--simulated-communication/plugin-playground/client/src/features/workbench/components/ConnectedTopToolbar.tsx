/**
 * TopToolbar wired to the RTK store.
 *
 * Derives plugin count, dispatch count, error count, and health status
 * from runtime + workbench state.
 */
import { useAppDispatch, useAppSelector, setActiveDevToolsTab, selectLoadedPluginIds, selectDispatchTimeline } from "@/store";
import { TopToolbar, type HealthStatus } from "./TopToolbar";

export function ConnectedTopToolbar() {
  const dispatch = useAppDispatch();

  const pluginIds = useAppSelector(selectLoadedPluginIds);
  const timeline = useAppSelector(selectDispatchTimeline);
  const errors = useAppSelector((s) => s.workbench.errors);

  // Derive health from error count
  let health: HealthStatus = "healthy";
  if (errors.length > 5) health = "error";
  else if (errors.length > 0) health = "degraded";

  return (
    <TopToolbar
      pluginCount={pluginIds.length}
      dispatchCount={timeline.length}
      health={health}
      errorCount={errors.length}
      onHealthClick={() => dispatch(setActiveDevToolsTab("errors"))}
    />
  );
}

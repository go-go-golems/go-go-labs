import React from "react";
import { cn } from "@/lib/utils";

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

export interface WorkbenchLayoutProps {
  /** Narrow left sidebar (catalog + running instances). */
  sidebar?: React.ReactNode;
  /** Top toolbar (badges, runtime status, menu). */
  toolbar?: React.ReactNode;
  /** Main content area — typically EditorTabBar + SplitView(editor, preview). */
  children: React.ReactNode;
  /** Bottom devtools panel (timeline, state, capabilities, errors, shared, docs). */
  devtools?: React.ReactNode;
  /** Whether the sidebar is collapsed (48px icon-only). */
  sidebarCollapsed?: boolean;
  /** Whether the devtools panel is collapsed (tab bar only). */
  devtoolsCollapsed?: boolean;
  className?: string;
}

// ---------------------------------------------------------------------------
// Component
// ---------------------------------------------------------------------------

export function WorkbenchLayout({
  sidebar,
  toolbar,
  children,
  devtools,
  sidebarCollapsed = false,
  devtoolsCollapsed = false,
  className,
}: WorkbenchLayoutProps) {
  return (
    <div
      data-widget="workbench"
      className={cn("h-dvh flex flex-col bg-background text-foreground overflow-hidden", className)}
    >
      {/* Top toolbar */}
      {toolbar && (
        <div data-part="toolbar" className="flex-shrink-0 border-b border-border">
          {toolbar}
        </div>
      )}

      {/* Main body: sidebar + center + (devtools is below center) */}
      <div className="flex flex-1 min-h-0">
        {/* Sidebar */}
        {sidebar && (
          <aside
            data-part="sidebar"
            data-state={sidebarCollapsed ? "collapsed" : "expanded"}
            className={cn(
              "flex-shrink-0 border-r border-border overflow-y-auto transition-[width] duration-150",
              sidebarCollapsed ? "w-12" : "w-60"
            )}
          >
            {sidebar}
          </aside>
        )}

        {/* Center: main + devtools stacked vertically */}
        <div className="flex flex-col flex-1 min-w-0 min-h-0">
          {/* Main content pane */}
          <main data-part="main" className="flex-1 min-h-0 overflow-hidden">
            {children}
          </main>

          {/* Devtools panel */}
          {devtools && (
            <div
              data-part="devtools"
              data-state={devtoolsCollapsed ? "collapsed" : "expanded"}
              className={cn(
                "flex-shrink-0 border-t border-border overflow-hidden transition-[height] duration-150",
                devtoolsCollapsed ? "h-9" : "h-70"
              )}
            >
              {devtools}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

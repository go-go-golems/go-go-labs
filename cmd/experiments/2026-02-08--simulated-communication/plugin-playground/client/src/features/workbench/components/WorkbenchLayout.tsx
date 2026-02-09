import React from "react";
import { cn } from "@/lib/utils";
import { useAppSelector } from "@/store";
import "../styles/workbench.css";

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

export interface WorkbenchLayoutProps {
  /** Narrow left sidebar (catalog + running instances). */
  sidebar?: React.ReactNode;
  /** Top toolbar (badges, runtime status, menu). */
  toolbar?: React.ReactNode;
  /** Main content area — EditorTabBar + SplitView(editor, preview). */
  children: React.ReactNode;
  /** Bottom devtools panel. */
  devtools?: React.ReactNode;
  /** Strip all workbench CSS; render only data-part attributes for consumer styling. */
  unstyled?: boolean;
  /** Override sidebar collapsed state (defaults to store). */
  sidebarCollapsed?: boolean;
  /** Override devtools collapsed state (defaults to store). */
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
  unstyled = false,
  sidebarCollapsed: sidebarCollapsedProp,
  devtoolsCollapsed: devtoolsCollapsedProp,
  className,
}: WorkbenchLayoutProps) {
  const storeCollapsed = useAppSelector((s) => s.workbench.sidebarCollapsed);
  const storeDevtools = useAppSelector((s) => s.workbench.devtoolsCollapsed);
  const sidebarCollapsed = sidebarCollapsedProp ?? storeCollapsed;
  const devtoolsCollapsed = devtoolsCollapsedProp ?? storeDevtools;

  return (
    <div
      data-widget="workbench"
      data-unstyled={unstyled || undefined}
      className={cn(!unstyled && "h-dvh", className)}
    >
      {toolbar && <div data-part="toolbar">{toolbar}</div>}

      <div data-part="body">
        {sidebar && (
          <aside
            data-part="sidebar"
            data-state={sidebarCollapsed ? "collapsed" : "expanded"}
          >
            {sidebar}
          </aside>
        )}

        <div data-part="center">
          <main data-part="main">{children}</main>

          {devtools && (
            <div
              data-part="devtools"
              data-state={devtoolsCollapsed ? "collapsed" : "expanded"}
            >
              {devtools}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

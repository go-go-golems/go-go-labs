import React, { useState, useCallback } from "react";
import { cn } from "@/lib/utils";
import { useAppSelector } from "@/store";
import { ResizeHandle } from "./ResizeHandle";
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
  /** Initial devtools height in px (default 280). */
  defaultDevtoolsHeight?: number;
  className?: string;
}

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

const DEFAULT_DEVTOOLS_HEIGHT = 280;
const MIN_DEVTOOLS_HEIGHT = 80;
const MAX_DEVTOOLS_HEIGHT = 600;
const COLLAPSED_DEVTOOLS_HEIGHT = 36; // just the tab bar

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
  defaultDevtoolsHeight = DEFAULT_DEVTOOLS_HEIGHT,
  className,
}: WorkbenchLayoutProps) {
  const storeCollapsed = useAppSelector((s) => s.workbench.sidebarCollapsed);
  const storeDevtools = useAppSelector((s) => s.workbench.devtoolsCollapsed);
  const sidebarCollapsed = sidebarCollapsedProp ?? storeCollapsed;
  const devtoolsCollapsed = devtoolsCollapsedProp ?? storeDevtools;

  const [devtoolsHeight, setDevtoolsHeight] = useState(defaultDevtoolsHeight);

  const handleResize = useCallback((newSize: number) => {
    setDevtoolsHeight(newSize);
  }, []);

  const effectiveHeight = devtoolsCollapsed ? COLLAPSED_DEVTOOLS_HEIGHT : devtoolsHeight;

  return (
    <div
      data-widget="workbench"
      data-unstyled={unstyled || undefined}
      className={cn(!unstyled && "h-dvh", className)}
      style={{
        "--wb-devtools-height": `${effectiveHeight}px`,
        "--wb-devtools-collapsed-height": `${COLLAPSED_DEVTOOLS_HEIGHT}px`,
      } as React.CSSProperties}
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
            <>
              {!devtoolsCollapsed && (
                <ResizeHandle
                  size={devtoolsHeight}
                  onResize={handleResize}
                  minSize={MIN_DEVTOOLS_HEIGHT}
                  maxSize={MAX_DEVTOOLS_HEIGHT}
                />
              )}
              <div
                data-part="devtools"
                data-state={devtoolsCollapsed ? "collapsed" : "expanded"}
                style={{ height: effectiveHeight }}
              >
                {devtools}
              </div>
            </>
          )}
        </div>
      </div>
    </div>
  );
}

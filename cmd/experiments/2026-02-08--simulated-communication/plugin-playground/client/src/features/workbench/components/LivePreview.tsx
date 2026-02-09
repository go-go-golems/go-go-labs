import React from "react";
import { cn } from "@/lib/utils";

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

export interface LivePreviewProps {
  /** Label displayed at the top. */
  label?: string;
  /** Instance cards rendered as children. */
  children?: React.ReactNode;
  className?: string;
}

// ---------------------------------------------------------------------------
// Component
// ---------------------------------------------------------------------------

export function LivePreview({
  label = "Live Preview",
  children,
  className,
}: LivePreviewProps) {
  const hasChildren = React.Children.count(children) > 0;

  return (
    <div
      data-part="live-preview"
      className={cn("flex flex-col h-full overflow-auto", className)}
    >
      {/* Header label */}
      <div
        data-part="live-preview-header"
        className="flex-shrink-0 px-4 pt-3 pb-2"
      >
        <span className="text-xs text-slate-500 uppercase tracking-wider font-medium">
          {label}
        </span>
      </div>

      {/* Cards area */}
      <div
        data-part="live-preview-body"
        className="flex-1 min-h-0 overflow-y-auto px-4 pb-4 space-y-3"
      >
        {hasChildren ? (
          children
        ) : (
          <div
            data-part="live-preview-empty"
            className="flex items-center justify-center h-full text-xs text-slate-600"
          >
            Load a plugin to see its live output here.
          </div>
        )}
      </div>
    </div>
  );
}

import React, { useCallback, useRef, useEffect } from "react";
import { cn } from "@/lib/utils";

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

export interface ResizeHandleProps {
  /** Current size in px of the panel being resized. */
  size: number;
  /** Called while dragging with the new size. */
  onResize: (newSize: number) => void;
  /** Called when drag ends. */
  onResizeEnd?: () => void;
  /** Minimum panel size in px. */
  minSize?: number;
  /** Maximum panel size in px. */
  maxSize?: number;
  /** Orientation — "horizontal" resizes height (drag up/down). */
  orientation?: "horizontal";
  className?: string;
}

// ---------------------------------------------------------------------------
// Component
// ---------------------------------------------------------------------------

export function ResizeHandle({
  size,
  onResize,
  onResizeEnd,
  minSize = 80,
  maxSize = 600,
  orientation = "horizontal",
  className,
}: ResizeHandleProps) {
  const dragging = useRef(false);
  const startY = useRef(0);
  const startSize = useRef(0);

  const handleMouseDown = useCallback(
    (e: React.MouseEvent) => {
      e.preventDefault();
      dragging.current = true;
      startY.current = e.clientY;
      startSize.current = size;
      document.body.style.cursor = "row-resize";
      document.body.style.userSelect = "none";
    },
    [size],
  );

  useEffect(() => {
    const handleMouseMove = (e: MouseEvent) => {
      if (!dragging.current) return;
      // Dragging up → increase panel height (mouse moves up, delta negative)
      const delta = startY.current - e.clientY;
      const newSize = Math.max(minSize, Math.min(maxSize, startSize.current + delta));
      onResize(newSize);
    };

    const handleMouseUp = () => {
      if (!dragging.current) return;
      dragging.current = false;
      document.body.style.cursor = "";
      document.body.style.userSelect = "";
      onResizeEnd?.();
    };

    document.addEventListener("mousemove", handleMouseMove);
    document.addEventListener("mouseup", handleMouseUp);
    return () => {
      document.removeEventListener("mousemove", handleMouseMove);
      document.removeEventListener("mouseup", handleMouseUp);
    };
  }, [onResize, onResizeEnd, minSize, maxSize]);

  return (
    <div
      data-part="resize-handle"
      onMouseDown={handleMouseDown}
      className={cn(
        "flex-shrink-0 h-1 cursor-row-resize group relative",
        "border-t border-white/[0.06]",
        className,
      )}
    >
      {/* Visual indicator — thin line that highlights on hover/drag */}
      <div
        className={cn(
          "absolute inset-x-0 -top-px h-[3px]",
          "bg-transparent group-hover:bg-blue-500/30 transition-colors",
        )}
      />
      {/* Wider invisible hit area */}
      <div className="absolute inset-x-0 -top-1 h-3" />
    </div>
  );
}

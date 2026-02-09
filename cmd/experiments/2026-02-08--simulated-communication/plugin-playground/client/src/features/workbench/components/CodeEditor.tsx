import React, { useCallback, useRef, useEffect } from "react";
import { cn } from "@/lib/utils";

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

export interface CodeEditorProps {
  /** Current source code. */
  value: string;
  /** Called on every edit. */
  onChange?: (value: string) => void;
  /** Language hint (for future syntax highlighting). */
  language?: "javascript" | "typescript";
  /** Make the editor read-only. */
  readOnly?: boolean;
  /** Placeholder when empty. */
  placeholder?: string;
  className?: string;
}

// ---------------------------------------------------------------------------
// Component
//
// Phase 1: simple <textarea> with monospace styling that matches vm-system-ui.
// Phase 2: replace with CodeMirror 6 (drop-in, same props).
// ---------------------------------------------------------------------------

export function CodeEditor({
  value,
  onChange,
  language = "javascript",
  readOnly = false,
  placeholder = "// Write your plugin code here…",
  className,
}: CodeEditorProps) {
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  // Handle Tab key → insert 2 spaces instead of focus-change
  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
      if (e.key === "Tab") {
        e.preventDefault();
        const ta = e.currentTarget;
        const start = ta.selectionStart;
        const end = ta.selectionEnd;
        const newValue = value.slice(0, start) + "  " + value.slice(end);
        onChange?.(newValue);
        // Restore cursor after React re-render
        requestAnimationFrame(() => {
          ta.selectionStart = ta.selectionEnd = start + 2;
        });
      }
    },
    [value, onChange],
  );

  // Auto-scroll to keep cursor visible
  useEffect(() => {
    if (textareaRef.current) {
      textareaRef.current.style.height = "auto";
      textareaRef.current.style.height = textareaRef.current.scrollHeight + "px";
    }
  }, [value]);

  return (
    <div
      data-part="code-editor"
      data-language={language}
      data-readonly={readOnly || undefined}
      className={cn("flex-1 min-h-0 overflow-auto bg-slate-950/50", className)}
    >
      <textarea
        ref={textareaRef}
        data-part="code-editor-input"
        value={value}
        onChange={(e) => onChange?.(e.target.value)}
        onKeyDown={handleKeyDown}
        readOnly={readOnly}
        placeholder={placeholder}
        spellCheck={false}
        autoCapitalize="off"
        autoComplete="off"
        autoCorrect="off"
        className={cn(
          "w-full h-full min-h-full resize-none p-4",
          "bg-transparent text-slate-300 placeholder:text-slate-700",
          "font-mono text-xs leading-relaxed",
          "outline-none border-none",
          "selection:bg-blue-500/30",
          readOnly && "cursor-default opacity-70",
        )}
      />
    </div>
  );
}

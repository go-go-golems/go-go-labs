import React, { useState, useMemo, useCallback } from "react";
import { cn } from "@/lib/utils";
import { Copy, Check, FileText } from "lucide-react";

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

export interface DocEntry {
  title: string;
  category: string;
  path: string;
  raw: string;
}

export interface DocsPanelProps {
  /** Doc entries to display. */
  docs: DocEntry[];
  /** Concatenated docs string for "Copy All". */
  allDocsMarkdown?: string;
  className?: string;
}

// ---------------------------------------------------------------------------
// Copy-to-clipboard helper
// ---------------------------------------------------------------------------

function useCopyFeedback() {
  const [copiedKey, setCopiedKey] = useState<string | null>(null);

  const copy = useCallback(async (text: string, key: string) => {
    await navigator.clipboard.writeText(text);
    setCopiedKey(key);
    setTimeout(() => setCopiedKey(null), 2000);
  }, []);

  return { copiedKey, copy };
}

// ---------------------------------------------------------------------------
// Component
// ---------------------------------------------------------------------------

export function DocsPanel({
  docs,
  allDocsMarkdown,
  className,
}: DocsPanelProps) {
  const [selectedPath, setSelectedPath] = useState(docs[0]?.path ?? "");
  const { copiedKey, copy } = useCopyFeedback();

  const selectedDoc = docs.find((d) => d.path === selectedPath);

  // Group docs by category
  const grouped = useMemo(() => {
    const map = new Map<string, DocEntry[]>();
    for (const doc of docs) {
      const arr = map.get(doc.category) ?? [];
      arr.push(doc);
      map.set(doc.category, arr);
    }
    return map;
  }, [docs]);

  if (docs.length === 0) {
    return (
      <div data-part="docs-panel" className={cn("flex items-center justify-center h-full text-xs text-slate-600", className)}>
        No documentation available.
      </div>
    );
  }

  return (
    <div data-part="docs-panel" className={cn("flex h-full", className)}>
      {/* Nav sidebar */}
      <div
        data-part="docs-nav"
        className="w-48 flex-shrink-0 border-r border-white/[0.06] overflow-y-auto flex flex-col"
      >
        <div className="flex-1 py-2">
          {Array.from(grouped.entries()).map(([category, entries]) => (
            <div key={category} className="mb-2">
              <div className="px-3 py-1 text-[10px] font-medium text-slate-600 uppercase tracking-wider">
                {category}
              </div>
              {entries.map((doc) => {
                const active = doc.path === selectedPath;
                return (
                  <button
                    key={doc.path}
                    data-part="docs-nav-item"
                    data-state={active ? "active" : undefined}
                    onClick={() => setSelectedPath(doc.path)}
                    className={cn(
                      "w-full text-left px-3 py-1 text-xs transition-colors truncate",
                      active
                        ? "bg-slate-800 text-slate-200"
                        : "text-slate-500 hover:text-slate-300 hover:bg-slate-800/50",
                    )}
                  >
                    <FileText className="w-3 h-3 inline-block mr-1.5 -mt-0.5" />
                    {doc.title}
                  </button>
                );
              })}
            </div>
          ))}
        </div>

        {/* Copy All button */}
        {allDocsMarkdown && (
          <div className="flex-shrink-0 border-t border-white/[0.06] p-2">
            <button
              data-part="docs-copy-all"
              onClick={() => copy(allDocsMarkdown, "all")}
              className="w-full flex items-center justify-center gap-1.5 h-7 rounded-md text-[10px] font-medium text-slate-500 hover:text-slate-200 border border-white/[0.08] hover:border-white/[0.15] hover:bg-slate-800/50 transition-colors"
            >
              {copiedKey === "all" ? (
                <>
                  <Check className="w-3 h-3 text-emerald-500" />
                  Copied!
                </>
              ) : (
                <>
                  <Copy className="w-3 h-3" />
                  Copy All Docs
                </>
              )}
            </button>
          </div>
        )}
      </div>

      {/* Content area */}
      <div data-part="docs-content" className="flex-1 min-w-0 flex flex-col overflow-hidden">
        {selectedDoc ? (
          <>
            {/* Doc header */}
            <div
              data-part="docs-content-header"
              className="flex items-center justify-between px-4 py-2 border-b border-white/[0.06] flex-shrink-0"
            >
              <div>
                <span className="text-xs font-medium text-slate-200">{selectedDoc.title}</span>
                <span className="ml-2 text-[10px] font-mono text-slate-600">{selectedDoc.path}</span>
              </div>
              <button
                data-part="docs-copy-doc"
                onClick={() => copy(selectedDoc.raw, selectedDoc.path)}
                className="flex items-center gap-1 px-2 py-1 rounded text-[10px] text-slate-500 hover:text-slate-200 hover:bg-slate-800/50 transition-colors"
                title="Copy this document's markdown"
              >
                {copiedKey === selectedDoc.path ? (
                  <>
                    <Check className="w-3 h-3 text-emerald-500" />
                    Copied
                  </>
                ) : (
                  <>
                    <Copy className="w-3 h-3" />
                    Copy
                  </>
                )}
              </button>
            </div>

            {/* Rendered markdown (Phase 1: pre-formatted; Phase 2: marked HTML) */}
            <div
              data-part="docs-rendered"
              className="flex-1 min-h-0 overflow-auto px-4 py-3"
            >
              <pre className="text-xs font-mono text-slate-300 whitespace-pre-wrap leading-relaxed">
                {selectedDoc.raw}
              </pre>
            </div>
          </>
        ) : (
          <div className="flex items-center justify-center h-full text-xs text-slate-600">
            Select a document.
          </div>
        )}
      </div>
    </div>
  );
}

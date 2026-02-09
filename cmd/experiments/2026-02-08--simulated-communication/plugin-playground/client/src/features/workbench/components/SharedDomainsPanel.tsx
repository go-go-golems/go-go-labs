import React from "react";
import { cn } from "@/lib/utils";

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

export interface SharedDomainInfo {
  name: string;
  state: unknown;
  readers: { instanceId: string; title: string; shortId: string }[];
  writers: { instanceId: string; title: string; shortId: string }[];
}

export interface SharedDomainsPanelProps {
  domains: SharedDomainInfo[];
  className?: string;
}

// ---------------------------------------------------------------------------
// Component
// ---------------------------------------------------------------------------

export function SharedDomainsPanel({
  domains,
  className,
}: SharedDomainsPanelProps) {
  if (domains.length === 0) {
    return (
      <div data-part="shared-panel" className={cn("flex items-center justify-center h-full text-xs text-slate-600", className)}>
        No shared domains registered.
      </div>
    );
  }

  return (
    <div data-part="shared-panel" className={cn("overflow-y-auto p-3 space-y-4", className)}>
      {domains.map((domain) => (
        <div
          key={domain.name}
          data-part="shared-domain-card"
          className="rounded-lg border border-white/[0.08] bg-slate-900/40"
        >
          {/* Domain header */}
          <div
            data-part="shared-domain-header"
            className="flex items-center justify-between px-3 py-2 border-b border-white/[0.06]"
          >
            <span className="text-xs font-medium text-slate-200 font-mono">{domain.name}</span>
            <div className="flex items-center gap-2 text-[10px] text-slate-600">
              <span>{domain.readers.length}R</span>
              <span>{domain.writers.length}W</span>
            </div>
          </div>

          {/* Participants */}
          <div className="px-3 py-2 flex gap-4 text-[10px] font-mono border-b border-white/[0.04]">
            <div>
              <span className="text-slate-500">Readers: </span>
              {domain.readers.length === 0 ? (
                <span className="text-slate-700">none</span>
              ) : (
                domain.readers.map((r, i) => (
                  <span key={r.instanceId}>
                    {i > 0 && ", "}
                    <span className="text-slate-400">{r.title}</span>
                    <span className="text-slate-600"> ({r.shortId})</span>
                  </span>
                ))
              )}
            </div>
            <div>
              <span className="text-slate-500">Writers: </span>
              {domain.writers.length === 0 ? (
                <span className="text-slate-700">none</span>
              ) : (
                domain.writers.map((w, i) => (
                  <span key={w.instanceId}>
                    {i > 0 && ", "}
                    <span className="text-slate-400">{w.title}</span>
                    <span className="text-slate-600"> ({w.shortId})</span>
                  </span>
                ))
              )}
            </div>
          </div>

          {/* State snapshot */}
          <pre
            data-part="shared-domain-state"
            className="px-3 py-2 text-xs font-mono text-slate-300 whitespace-pre-wrap overflow-auto max-h-32"
          >
            {JSON.stringify(domain.state, null, 2)}
          </pre>
        </div>
      ))}
    </div>
  );
}

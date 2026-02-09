import React, { useMemo, useState } from "react";
import { cn } from "@/lib/utils";

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

export type DispatchScope = "plugin" | "shared";
export type DispatchOutcome = "applied" | "denied" | "ignored";

export interface TimelineEntry {
  id: string;
  timestamp: number;
  scope: DispatchScope;
  outcome: DispatchOutcome;
  actionType: string;
  instanceId: string;
  shortInstanceId: string;
  domain?: string;
  reason?: string;
}

export interface TimelinePanelProps {
  entries: TimelineEntry[];
  /** Instance ID to highlight rows for. */
  focusedInstanceId?: string | null;
  className?: string;
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function formatTime(ts: number): string {
  const d = new Date(ts);
  return d.toLocaleTimeString("en-US", { hour12: false, fractionalSecondDigits: 1 });
}

const OUTCOME_STYLES: Record<DispatchOutcome, string> = {
  applied: "text-emerald-500",
  denied: "text-red-400",
  ignored: "text-slate-500",
};

// ---------------------------------------------------------------------------
// Component
// ---------------------------------------------------------------------------

export function TimelinePanel({
  entries,
  focusedInstanceId,
  className,
}: TimelinePanelProps) {
  const [scopeFilter, setScopeFilter] = useState<DispatchScope | "all">("all");
  const [outcomeFilter, setOutcomeFilter] = useState<DispatchOutcome | "all">("all");

  const filtered = useMemo(() => {
    let result = entries;
    if (scopeFilter !== "all") result = result.filter((e) => e.scope === scopeFilter);
    if (outcomeFilter !== "all") result = result.filter((e) => e.outcome === outcomeFilter);
    return result;
  }, [entries, scopeFilter, outcomeFilter]);

  return (
    <div data-part="timeline-panel" className={cn("flex flex-col h-full", className)}>
      {/* Filter bar */}
      <div data-part="timeline-filters" className="flex items-center gap-3 px-3 py-1.5 border-b border-white/[0.06] flex-shrink-0">
        <FilterSelect
          label="Scope"
          value={scopeFilter}
          options={["all", "plugin", "shared"]}
          onChange={(v) => setScopeFilter(v as any)}
        />
        <FilterSelect
          label="Outcome"
          value={outcomeFilter}
          options={["all", "applied", "denied", "ignored"]}
          onChange={(v) => setOutcomeFilter(v as any)}
        />
        <span className="ml-auto text-[10px] text-slate-600 tabular-nums">{filtered.length} / {entries.length}</span>
      </div>

      {/* Table */}
      <div data-part="timeline-body" className="flex-1 min-h-0 overflow-auto px-3 py-1">
        {filtered.length === 0 ? (
          <div className="flex items-center justify-center h-full text-xs text-slate-600">
            {entries.length === 0 ? "No dispatches yet." : "No matching dispatches."}
          </div>
        ) : (
          <table className="w-full text-left text-xs font-mono">
            <thead>
              <tr className="text-slate-600">
                <th className="pb-1 pr-3 font-medium">Time</th>
                <th className="pb-1 pr-3 font-medium">Scope</th>
                <th className="pb-1 pr-3 font-medium">Outcome</th>
                <th className="pb-1 pr-3 font-medium">Action</th>
                <th className="pb-1 pr-3 font-medium">Domain</th>
                <th className="pb-1 font-medium">Instance</th>
              </tr>
            </thead>
            <tbody className="text-slate-400">
              {filtered.map((entry) => (
                <tr
                  key={entry.id}
                  data-part="timeline-row"
                  data-state={focusedInstanceId === entry.instanceId ? "focused" : undefined}
                  className={cn(
                    "hover:bg-slate-800/30",
                    focusedInstanceId === entry.instanceId && "bg-blue-500/5",
                  )}
                >
                  <td className="py-0.5 pr-3 text-slate-600 whitespace-nowrap">{formatTime(entry.timestamp)}</td>
                  <td className="py-0.5 pr-3 whitespace-nowrap">{entry.scope}</td>
                  <td className={cn("py-0.5 pr-3 whitespace-nowrap", OUTCOME_STYLES[entry.outcome])}>{entry.outcome}</td>
                  <td className="py-0.5 pr-3 text-slate-300 whitespace-nowrap">{entry.actionType}</td>
                  <td className="py-0.5 pr-3 text-slate-600 whitespace-nowrap">{entry.domain ?? "—"}</td>
                  <td className="py-0.5 text-slate-600 whitespace-nowrap">{entry.shortInstanceId}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
}

// ---------------------------------------------------------------------------
// Filter select helper
// ---------------------------------------------------------------------------

function FilterSelect({
  label,
  value,
  options,
  onChange,
}: {
  label: string;
  value: string;
  options: string[];
  onChange: (value: string) => void;
}) {
  return (
    <label className="flex items-center gap-1 text-[10px] text-slate-500">
      {label}:
      <select
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className="bg-slate-800 border border-white/[0.08] rounded px-1.5 py-0.5 text-slate-300 text-[10px] outline-none"
      >
        {options.map((o) => (
          <option key={o} value={o}>{o}</option>
        ))}
      </select>
    </label>
  );
}

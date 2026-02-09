import React from "react";
import { cn } from "@/lib/utils";
import { Check, X } from "lucide-react";

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

export interface InstanceCapabilities {
  instanceId: string;
  title: string;
  shortId: string;
  grants: Record<string, { read: boolean; write: boolean }>;
}

export interface CapabilitiesPanelProps {
  /** All known shared domain names. */
  domains: string[];
  /** Per-instance grant info. */
  instances: InstanceCapabilities[];
  focusedInstanceId?: string | null;
  className?: string;
}

// ---------------------------------------------------------------------------
// Component
// ---------------------------------------------------------------------------

export function CapabilitiesPanel({
  domains,
  instances,
  focusedInstanceId,
  className,
}: CapabilitiesPanelProps) {
  if (instances.length === 0) {
    return (
      <div data-part="capabilities-panel" className={cn("flex items-center justify-center h-full text-xs text-slate-600", className)}>
        No instances loaded.
      </div>
    );
  }

  return (
    <div data-part="capabilities-panel" className={cn("overflow-auto p-3", className)}>
      <table className="w-full text-xs font-mono text-left">
        <thead>
          <tr className="text-slate-600">
            <th className="pb-2 pr-4 font-medium sticky left-0 bg-slate-900 z-10">Instance</th>
            {domains.map((d) => (
              <th key={d} className="pb-2 px-2 font-medium text-center whitespace-nowrap" colSpan={2}>
                {d}
              </th>
            ))}
          </tr>
          <tr className="text-slate-700">
            <th className="pb-1 pr-4 sticky left-0 bg-slate-900 z-10" />
            {domains.map((d) => (
              <React.Fragment key={d}>
                <th className="pb-1 px-2 font-normal text-center text-[10px]">R</th>
                <th className="pb-1 px-2 font-normal text-center text-[10px]">W</th>
              </React.Fragment>
            ))}
          </tr>
        </thead>
        <tbody>
          {instances.map((inst) => (
            <tr
              key={inst.instanceId}
              data-part="capability-row"
              data-state={focusedInstanceId === inst.instanceId ? "focused" : undefined}
              className={cn(
                "hover:bg-slate-800/30",
                focusedInstanceId === inst.instanceId && "bg-blue-500/5",
              )}
            >
              <td className="py-1 pr-4 text-slate-300 whitespace-nowrap sticky left-0 bg-slate-900 z-10">
                <span>{inst.title}</span>
                <span className="ml-1.5 text-slate-600">{inst.shortId}</span>
              </td>
              {domains.map((d) => {
                const g = inst.grants[d];
                return (
                  <React.Fragment key={d}>
                    <td className="py-1 px-2 text-center">
                      <GrantIcon granted={g?.read ?? false} />
                    </td>
                    <td className="py-1 px-2 text-center">
                      <GrantIcon granted={g?.write ?? false} />
                    </td>
                  </React.Fragment>
                );
              })}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function GrantIcon({ granted }: { granted: boolean }) {
  return granted ? (
    <Check className="w-3 h-3 text-emerald-500 inline-block" />
  ) : (
    <X className="w-3 h-3 text-slate-700 inline-block" />
  );
}

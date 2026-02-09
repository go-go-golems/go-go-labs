// Design Philosophy: Technical Brutalism - Interpret data-only UI trees into React
// Monospace typography, high contrast, glowing borders on interactive elements

import React from "react";
import type { UINode, UIEventRef } from "@runtime/uiTypes";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";

interface WidgetRendererProps {
  tree: UINode | null;
  onEvent: (ref: UIEventRef, eventPayload?: any) => void;
}

export function WidgetRenderer({ tree, onEvent }: WidgetRendererProps) {
  if (!tree) return <div className="text-muted-foreground font-mono text-sm">No widget tree</div>;

  return <>{renderNode(tree, onEvent)}</>;
}

function renderNode(node: UINode, onEvent: (ref: UIEventRef, eventPayload?: any) => void): React.ReactNode {
  if (!node) return null;

  switch (node.kind) {
    case "panel":
      return (
        <div 
          className="border border-accent/30 rounded-sm p-4 mb-3 bg-card/50 shadow-[0_0_15px_rgba(0,255,255,0.1)]"
          style={{ fontFamily: "'Space Mono', monospace" }}
        >
          {(node.children ?? []).map((c, i) => (
            <React.Fragment key={i}>{renderNode(c, onEvent)}</React.Fragment>
          ))}
        </div>
      );

    case "row":
      return (
        <div className="flex gap-3 items-center mb-3">
          {(node.children ?? []).map((c, i) => (
            <React.Fragment key={i}>{renderNode(c, onEvent)}</React.Fragment>
          ))}
        </div>
      );

    case "column":
      return (
        <div className="flex flex-col gap-2">
          {(node.children ?? []).map((c, i) => (
            <React.Fragment key={i}>{renderNode(c, onEvent)}</React.Fragment>
          ))}
        </div>
      );

    case "text":
      return <div className="text-foreground font-mono text-sm mb-2">{node.text}</div>;

    case "badge":
      return (
        <Badge 
          variant="outline" 
          className="font-mono text-xs uppercase tracking-wider border-accent/50 text-accent shadow-[0_0_8px_rgba(0,255,255,0.3)]"
        >
          {node.text}
        </Badge>
      );

    case "button": {
      const { label, onClick, variant } = node.props;
      return (
        <Button
          onClick={() => {
            onClick && onEvent(onClick, onClick.args);
          }}
          variant={variant === "destructive" ? "destructive" : "outline"}
          className="font-mono text-xs uppercase tracking-wide transition-all duration-200 hover:shadow-[0_0_12px_rgba(0,255,255,0.4)] border-accent/50"
        >
          {label}
        </Button>
      );
    }

    case "input": {
      const { value, placeholder, onChange } = node.props;
      return (
        <Input
          value={value}
          placeholder={placeholder}
          onChange={(e) => onChange && onEvent(onChange, { value: e.target.value })}
          className="font-mono text-sm border-accent/30 focus:border-accent focus:shadow-[0_0_10px_rgba(0,255,255,0.3)] transition-all"
        />
      );
    }

    case "counter": {
      const { value, onIncrement, onDecrement } = node.props;
      return (
        <div className="flex items-center gap-3 border border-accent/30 rounded-sm p-2 bg-card/30">
          <Button
            onClick={() => onDecrement && onEvent(onDecrement)}
            variant="outline"
            size="sm"
            className="font-mono border-accent/50 hover:shadow-[0_0_10px_rgba(0,255,255,0.4)]"
          >
            −
          </Button>
          <span className="font-mono text-lg font-bold text-accent min-w-[3ch] text-center">
            {value}
          </span>
          <Button
            onClick={() => onIncrement && onEvent(onIncrement)}
            variant="outline"
            size="sm"
            className="font-mono border-accent/50 hover:shadow-[0_0_10px_rgba(0,255,255,0.4)]"
          >
            +
          </Button>
        </div>
      );
    }

    case "table": {
      const { headers, rows } = node.props;
      return (
        <div className="border border-accent/30 rounded-sm overflow-hidden">
          <table className="w-full font-mono text-sm">
            <thead className="bg-accent/10 border-b border-accent/30">
              <tr>
                {headers.map((h: string, i: number) => (
                  <th key={i} className="text-left p-2 font-bold uppercase tracking-wide text-xs text-accent">
                    {h}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody>
              {rows.map((row: any[], i: number) => (
                <tr key={i} className="border-b border-accent/10 hover:bg-accent/5 transition-colors">
                  {row.map((cell, j) => (
                    <td key={j} className="p-2 text-foreground">
                      {String(cell)}
                    </td>
                  ))}
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      );
    }

    default:
      return null;
  }
}

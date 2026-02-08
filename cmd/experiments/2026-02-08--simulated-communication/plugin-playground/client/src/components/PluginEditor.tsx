// Design Philosophy: Technical Brutalism - Raw code editor with monospace aesthetic
// Monaco editor with dark theme matching terminal aesthetic

import React from "react";
import Editor from "@monaco-editor/react";
import { Button } from "@/components/ui/button";
import { Play, Save, X } from "lucide-react";

interface PluginEditorProps {
  code: string;
  onChange: (code: string) => void;
  onRun: () => void;
  onClose?: () => void;
  readOnly?: boolean;
}

export function PluginEditor({ code, onChange, onRun, onClose, readOnly = false }: PluginEditorProps) {
  return (
    <div className="flex flex-col h-full border border-accent/30 rounded-sm overflow-hidden bg-card shadow-[0_0_20px_rgba(0,255,255,0.1)]">
      <div className="flex items-center justify-between px-4 py-2 border-b border-accent/30 bg-accent/5">
        <div className="flex items-center gap-2">
          <span className="font-mono text-xs uppercase tracking-wider text-accent font-bold">
            Plugin Editor
          </span>
          <span className="font-mono text-xs text-muted-foreground">
            [Unified Runtime]
          </span>
        </div>
        <div className="flex items-center gap-2">
          {!readOnly && (
            <Button
              onClick={onRun}
              size="sm"
              variant="outline"
              className="font-mono text-xs uppercase tracking-wide border-accent/50 hover:shadow-[0_0_10px_rgba(0,255,255,0.4)]"
            >
              <Play className="w-3 h-3 mr-1" />
              Load Plugin
            </Button>
          )}
          {onClose && (
            <Button
              onClick={onClose}
              size="sm"
              variant="ghost"
              className="font-mono text-xs"
            >
              <X className="w-4 h-4" />
            </Button>
          )}
        </div>
      </div>
      
      <div className="flex-1 min-h-0">
        <Editor
          height="100%"
          defaultLanguage="javascript"
          value={code}
          onChange={(value) => onChange(value || "")}
          theme="vs-dark"
          options={{
            minimap: { enabled: false },
            fontSize: 13,
            fontFamily: "'JetBrains Mono', 'Courier New', monospace",
            lineNumbers: "on",
            scrollBeyondLastLine: false,
            automaticLayout: true,
            tabSize: 2,
            readOnly,
            wordWrap: "on",
            padding: { top: 16, bottom: 16 },
          }}
        />
      </div>
    </div>
  );
}

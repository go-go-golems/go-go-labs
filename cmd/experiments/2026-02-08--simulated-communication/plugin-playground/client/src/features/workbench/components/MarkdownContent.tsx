import React, { useCallback, useState } from "react";
import Markdown from "react-markdown";
import remarkGfm from "remark-gfm";
import rehypeHighlight from "rehype-highlight";
import { cn } from "@/lib/utils";
import { Copy, Check } from "lucide-react";

// highlight.js dark theme — atom-one-dark matches the slate palette well
import "highlight.js/styles/atom-one-dark.min.css";

// ---------------------------------------------------------------------------
// Code block with copy button
// ---------------------------------------------------------------------------

function CodeBlock({
  children,
  className,
  ...props
}: React.HTMLAttributes<HTMLElement> & { children?: React.ReactNode }) {
  const [copied, setCopied] = useState(false);

  // Extract raw text from children for clipboard
  const rawText = extractText(children);

  const handleCopy = useCallback(async () => {
    await navigator.clipboard.writeText(rawText);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  }, [rawText]);

  // Detect language from className (e.g. "hljs language-js")
  const lang = className
    ?.split(" ")
    .find((c) => c.startsWith("language-"))
    ?.replace("language-", "");

  return (
    <div data-part="md-code-block" className="relative group my-3 rounded-lg overflow-hidden border border-white/[0.06]">
      {/* Header with language + copy button */}
      <div className="flex items-center justify-between px-3 py-1 bg-slate-800/80 border-b border-white/[0.04]">
        <span className="text-[10px] font-mono text-slate-500 uppercase">{lang ?? "code"}</span>
        <button
          onClick={handleCopy}
          className="flex items-center gap-1 px-1.5 py-0.5 rounded text-[10px] text-slate-500 hover:text-slate-200 hover:bg-slate-700/50 transition-colors opacity-0 group-hover:opacity-100"
        >
          {copied ? (
            <><Check className="w-3 h-3 text-emerald-500" />Copied</>
          ) : (
            <><Copy className="w-3 h-3" />Copy</>
          )}
        </button>
      </div>
      <pre className="!m-0 !rounded-none !border-0 bg-slate-950/60 overflow-x-auto">
        <code className={cn(className, "!bg-transparent text-xs leading-relaxed")} {...props}>
          {children}
        </code>
      </pre>
    </div>
  );
}

/** Recursively extract text content from React children. */
function extractText(node: React.ReactNode): string {
  if (typeof node === "string") return node;
  if (typeof node === "number") return String(node);
  if (Array.isArray(node)) return node.map(extractText).join("");
  if (React.isValidElement(node) && node.props) {
    return extractText((node.props as { children?: React.ReactNode }).children);
  }
  return "";
}

// ---------------------------------------------------------------------------
// Inline code
// ---------------------------------------------------------------------------

function InlineCode({ children, ...props }: React.HTMLAttributes<HTMLElement>) {
  return (
    <code
      className="px-1.5 py-0.5 rounded bg-slate-800 text-blue-400 text-[0.85em] font-mono"
      {...props}
    >
      {children}
    </code>
  );
}

// ---------------------------------------------------------------------------
// Custom component map
// ---------------------------------------------------------------------------

const components = {
  // Override code rendering: fenced blocks → CodeBlock, inline → InlineCode
  code({ className, children, ...props }: any) {
    // react-markdown wraps fenced code blocks in <pre><code>; inline code is just <code>.
    // We detect fenced blocks by the presence of a className (set by rehype-highlight).
    const isBlock = className?.includes("hljs") || className?.includes("language-");
    if (isBlock) {
      return <CodeBlock className={className} {...props}>{children}</CodeBlock>;
    }
    return <InlineCode {...props}>{children}</InlineCode>;
  },
  // Wrap <pre> as pass-through so CodeBlock handles all styling
  pre({ children }: any) {
    return <>{children}</>;
  },
  // Tables
  table({ children }: any) {
    return (
      <div className="my-3 overflow-x-auto rounded-lg border border-white/[0.06]">
        <table className="w-full text-xs">{children}</table>
      </div>
    );
  },
  thead({ children }: any) {
    return <thead className="bg-slate-800/50 border-b border-white/[0.06]">{children}</thead>;
  },
  th({ children }: any) {
    return <th className="px-3 py-2 text-left font-medium text-slate-400 text-[10px] uppercase tracking-wider">{children}</th>;
  },
  td({ children }: any) {
    return <td className="px-3 py-1.5 text-slate-300 border-t border-white/[0.03]">{children}</td>;
  },
  // Links
  a({ href, children }: any) {
    return (
      <a href={href} className="text-blue-400 hover:text-blue-300 underline underline-offset-2 transition-colors" target="_blank" rel="noopener">
        {children}
      </a>
    );
  },
  // Block-level overrides for spacing + typography
  h1({ children }: any) {
    return <h1 className="text-base font-semibold text-slate-100 mt-6 mb-3 first:mt-0">{children}</h1>;
  },
  h2({ children }: any) {
    return <h2 className="text-sm font-semibold text-slate-200 mt-5 mb-2 border-b border-white/[0.06] pb-1">{children}</h2>;
  },
  h3({ children }: any) {
    return <h3 className="text-sm font-medium text-slate-300 mt-4 mb-1.5">{children}</h3>;
  },
  p({ children }: any) {
    return <p className="text-xs text-slate-400 leading-relaxed mb-2">{children}</p>;
  },
  ul({ children }: any) {
    return <ul className="text-xs text-slate-400 list-disc list-inside mb-2 space-y-0.5">{children}</ul>;
  },
  ol({ children }: any) {
    return <ol className="text-xs text-slate-400 list-decimal list-inside mb-2 space-y-0.5">{children}</ol>;
  },
  li({ children }: any) {
    return <li className="leading-relaxed">{children}</li>;
  },
  blockquote({ children }: any) {
    return <blockquote className="border-l-2 border-blue-500/30 pl-3 my-2 text-xs text-slate-500 italic">{children}</blockquote>;
  },
  hr() {
    return <hr className="border-white/[0.06] my-4" />;
  },
  strong({ children }: any) {
    return <strong className="font-semibold text-slate-200">{children}</strong>;
  },
};

// ---------------------------------------------------------------------------
// Main component
// ---------------------------------------------------------------------------

export interface MarkdownContentProps {
  /** Raw markdown string. */
  source: string;
  className?: string;
}

export function MarkdownContent({ source, className }: MarkdownContentProps) {
  return (
    <div data-part="md-content" className={cn("prose-none", className)}>
      <Markdown
        remarkPlugins={[remarkGfm]}
        rehypePlugins={[rehypeHighlight]}
        components={components}
      >
        {source}
      </Markdown>
    </div>
  );
}

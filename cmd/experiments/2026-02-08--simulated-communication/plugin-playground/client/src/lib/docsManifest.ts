/**
 * Documentation manifest — raw markdown embedded at build time via Vite ?raw.
 *
 * The docs are bundled into the JS output so the DocsPanel works without
 * any server-side file access. Bundle cost: ~15KB gzipped for the current
 * seven docs.
 */

import readmeRaw from "../../../docs/README.md?raw";
import quickstartRaw from "../../../docs/plugin-authoring/quickstart.md?raw";
import examplesRaw from "../../../docs/plugin-authoring/examples.md?raw";
import dispatchLifecycleRaw from "../../../docs/architecture/dispatch-lifecycle.md?raw";
import uiDslRaw from "../../../docs/architecture/ui-dsl.md?raw";
import capabilityModelRaw from "../../../docs/architecture/capability-model.md?raw";
import embeddingRaw from "../../../docs/runtime/embedding.md?raw";
import changelogRaw from "../../../docs/migration/changelog-vm-api.md?raw";

export interface DocEntry {
  /** Display title in nav tree. */
  title: string;
  /** Category / parent folder for nav grouping. */
  category: string;
  /** Relative path from docs/ root (for display and "Copy All" headers). */
  path: string;
  /** Raw markdown source (for copy-to-clipboard). */
  raw: string;
}

export const docs: DocEntry[] = [
  {
    title: "Overview",
    category: "Overview",
    path: "docs/README.md",
    raw: readmeRaw,
  },
  {
    title: "Plugin Authoring Guide",
    category: "Plugin Authoring",
    path: "docs/plugin-authoring/quickstart.md",
    raw: quickstartRaw,
  },
  {
    title: "Plugin Examples",
    category: "Plugin Authoring",
    path: "docs/plugin-authoring/examples.md",
    raw: examplesRaw,
  },
  {
    title: "Dispatch Lifecycle",
    category: "Architecture",
    path: "docs/architecture/dispatch-lifecycle.md",
    raw: dispatchLifecycleRaw,
  },
  {
    title: "UI DSL Reference",
    category: "Architecture",
    path: "docs/architecture/ui-dsl.md",
    raw: uiDslRaw,
  },
  {
    title: "Capability Model",
    category: "Architecture",
    path: "docs/architecture/capability-model.md",
    raw: capabilityModelRaw,
  },
  {
    title: "Embedding Guide",
    category: "Runtime",
    path: "docs/runtime/embedding.md",
    raw: embeddingRaw,
  },
  {
    title: "VM API Changelog",
    category: "Migration",
    path: "docs/migration/changelog-vm-api.md",
    raw: changelogRaw,
  },
];

/**
 * Build the concatenated "all docs" string for the Copy All button.
 */
export function buildAllDocsMarkdown(): string {
  const parts = docs.map((d) => `# ${d.path}\n\n${d.raw}`);
  return `# Plugin Playground Documentation\n\n---\n\n${parts.join("\n\n---\n\n")}`;
}

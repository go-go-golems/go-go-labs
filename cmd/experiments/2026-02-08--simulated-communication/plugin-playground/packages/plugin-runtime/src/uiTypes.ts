// Design Philosophy: Technical Brutalism - Data-only widget definitions
// No React components in plugins, only JSON-serializable UI trees

export type UIEventRef = { handler: string; args?: any };

export type UINode =
  | { kind: "panel" | "row" | "column"; props?: any; children?: UINode[] }
  | { kind: "text" | "badge"; props?: any; text: string }
  | { kind: "button"; props: { label: string; onClick?: UIEventRef; variant?: string } }
  | { kind: "input"; props: { value: string; placeholder?: string; onChange?: UIEventRef } }
  | { kind: "counter"; props: { value: number; onIncrement?: UIEventRef; onDecrement?: UIEventRef } }
  | { kind: "table"; props: { headers: string[]; rows: any[][] } };

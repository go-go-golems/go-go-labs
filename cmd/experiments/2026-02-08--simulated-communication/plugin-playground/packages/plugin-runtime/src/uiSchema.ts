import type { UINode, UIEventRef } from "./uiTypes";

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function assertEventRef(value: unknown, path: string): asserts value is UIEventRef {
  if (!isRecord(value)) {
    throw new Error(`${path} must be an object`);
  }
  if (typeof value.handler !== "string" || value.handler.length === 0) {
    throw new Error(`${path}.handler must be a non-empty string`);
  }
}

export function assertUINode(value: unknown, path = "root"): asserts value is UINode {
  if (!isRecord(value)) {
    throw new Error(`${path} must be an object`);
  }

  const kind = value.kind;
  if (typeof kind !== "string") {
    throw new Error(`${path}.kind must be a string`);
  }

  if (kind === "panel" || kind === "row" || kind === "column") {
    if (value.children !== undefined) {
      if (!Array.isArray(value.children)) {
        throw new Error(`${path}.children must be an array`);
      }
      value.children.forEach((child, index) => assertUINode(child, `${path}.children[${index}]`));
    }
    return;
  }

  if (kind === "text" || kind === "badge") {
    if (typeof value.text !== "string") {
      throw new Error(`${path}.text must be a string`);
    }
    return;
  }

  if (kind === "button") {
    if (!isRecord(value.props) || typeof value.props.label !== "string") {
      throw new Error(`${path}.props.label must be a string`);
    }
    if (value.props.onClick !== undefined) {
      assertEventRef(value.props.onClick, `${path}.props.onClick`);
    }
    return;
  }

  if (kind === "input") {
    if (!isRecord(value.props) || typeof value.props.value !== "string") {
      throw new Error(`${path}.props.value must be a string`);
    }
    if (value.props.onChange !== undefined) {
      assertEventRef(value.props.onChange, `${path}.props.onChange`);
    }
    return;
  }

  if (kind === "counter") {
    if (!isRecord(value.props) || typeof value.props.value !== "number") {
      throw new Error(`${path}.props.value must be a number`);
    }
    if (value.props.onIncrement !== undefined) {
      assertEventRef(value.props.onIncrement, `${path}.props.onIncrement`);
    }
    if (value.props.onDecrement !== undefined) {
      assertEventRef(value.props.onDecrement, `${path}.props.onDecrement`);
    }
    return;
  }

  if (kind === "table") {
    if (!isRecord(value.props)) {
      throw new Error(`${path}.props must be an object`);
    }
    if (!Array.isArray(value.props.headers) || value.props.headers.some((h) => typeof h !== "string")) {
      throw new Error(`${path}.props.headers must be a string[]`);
    }
    if (
      !Array.isArray(value.props.rows) ||
      value.props.rows.some((row) => !Array.isArray(row))
    ) {
      throw new Error(`${path}.props.rows must be an array of rows`);
    }
    return;
  }

  throw new Error(`${path}.kind '${kind}' is not supported`);
}

export function validateUINode(value: unknown): UINode {
  assertUINode(value);
  return value;
}


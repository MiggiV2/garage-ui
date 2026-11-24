import { reactive } from "vue";

export interface Toast {
  id: number;
  kind: "success" | "error";
  text: string;
}

let nextId = 1;
export const toasts = reactive<Toast[]>([]);

export function notify(kind: Toast["kind"], text: string) {
  const id = nextId++;
  toasts.push({ id, kind, text });
  setTimeout(() => dismiss(id), kind === "error" ? 8000 : 3500);
}

export function dismiss(id: number) {
  const index = toasts.findIndex((t) => t.id === id);
  if (index >= 0) toasts.splice(index, 1);
}

export function errorText(error: unknown): string {
  return error instanceof Error ? error.message : String(error);
}

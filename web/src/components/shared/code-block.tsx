import { CopyButton } from "./copy-button";

export function CodeBlock({ code, language = "bash" }: { code: string; language?: string }) {
  return (
    <div className="relative rounded-lg bg-zinc-900 border border-border p-4 font-mono text-sm">
      <CopyButton value={code} className="absolute top-2 right-2" />
      <pre className="overflow-x-auto text-zinc-300">{code}</pre>
    </div>
  );
}

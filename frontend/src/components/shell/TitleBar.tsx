import { useEffect, useState, type CSSProperties, type ReactNode } from "react";
import { Window } from "@wailsio/runtime";
import { GetVersion } from "../../../bindings/github.com/its-haze/league-rpc/cmd/league-rpc-gui/guiservice";

// Custom chrome for the frameless window (see Frameless in main.go).
export function TitleBar() {
  const [version, setVersion] = useState("");
  const [maximised, setMaximised] = useState(false);

  useEffect(() => {
    GetVersion().then(setVersion).catch(() => {});
  }, []);

  useEffect(() => {
    Window.IsMaximised()
      .then(setMaximised)
      .catch(() => {});
  }, []);

  async function toggleMaximise() {
    await Window.ToggleMaximise();
    setMaximised(await Window.IsMaximised().catch(() => !maximised));
  }

  return (
    <div
      className="border-border bg-surface flex h-9 shrink-0 items-center justify-between border-b select-none"
      style={{ "--wails-draggable": "drag" } as CSSProperties}
    >
      <div className="flex items-center gap-2 px-3">
        <span className="text-xs font-bold tracking-wider uppercase">League RPC</span>
        {version && <span className="text-muted text-[11px]">v{version}</span>}
      </div>

      <div className="flex h-full items-stretch" style={{ "--wails-draggable": "no-drag" } as CSSProperties}>
        <TitleBarButton label="Minimize" onClick={() => void Window.Minimise()}>
          <svg viewBox="0 0 10 10" width="10" height="10" aria-hidden>
            <rect x="0" y="4.5" width="10" height="1" fill="currentColor" />
          </svg>
        </TitleBarButton>
        <TitleBarButton label={maximised ? "Restore" : "Maximize"} onClick={() => void toggleMaximise()}>
          {maximised ? (
            <svg viewBox="0 0 10 10" width="10" height="10" aria-hidden>
              <path
                d="M2 0h8v8h-2M0 2h8v8H0z"
                fill="none"
                stroke="currentColor"
                strokeWidth="1"
              />
            </svg>
          ) : (
            <svg viewBox="0 0 10 10" width="10" height="10" aria-hidden>
              <rect x="0.5" y="0.5" width="9" height="9" fill="none" stroke="currentColor" strokeWidth="1" />
            </svg>
          )}
        </TitleBarButton>
        <TitleBarButton label="Close" danger onClick={() => void Window.Close()}>
          <svg viewBox="0 0 10 10" width="10" height="10" aria-hidden>
            <path d="M0 0l10 10M10 0L0 10" stroke="currentColor" strokeWidth="1" />
          </svg>
        </TitleBarButton>
      </div>
    </div>
  );
}

function TitleBarButton({
  label,
  danger,
  onClick,
  children,
}: {
  label: string;
  danger?: boolean;
  onClick: () => void;
  children: ReactNode;
}) {
  return (
    <button
      onClick={onClick}
      aria-label={label}
      title={label}
      className={
        "text-muted flex w-11 items-center justify-center transition-colors " +
        (danger ? "hover:bg-danger hover:text-danger-text" : "hover:bg-surface-raised hover:text-text")
      }
    >
      {children}
    </button>
  );
}

import * as SwitchPrimitive from "@radix-ui/react-switch";

export interface ToggleProps {
  checked: boolean;
  onCheckedChange: (checked: boolean) => void;
  disabled?: boolean;
  label?: string;
  id?: string;
}

// The on/off primitive every boolean setting (pause, show-rank, launch-at-
export function Toggle({ checked, onCheckedChange, disabled, label, id }: ToggleProps) {
  return (
    <SwitchPrimitive.Root
      id={id}
      checked={checked}
      onCheckedChange={onCheckedChange}
      disabled={disabled}
      aria-label={label}
      className="border-border bg-surface-raised data-[state=checked]:bg-accent relative h-5 w-9 rounded-full border transition-colors disabled:opacity-60"
    >
      <SwitchPrimitive.Thumb className="bg-surface data-[state=checked]:bg-accent-text block size-3.5 translate-x-0.5 rounded-full transition-transform data-[state=checked]:translate-x-[1.15rem]" />
    </SwitchPrimitive.Root>
  );
}

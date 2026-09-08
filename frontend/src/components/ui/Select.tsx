import * as SelectPrimitive from "@radix-ui/react-select";

export interface SelectOption {
  value: string;
  label: string;
}

export interface SelectProps {
  value: string;
  onValueChange: (value: string) => void;
  options: SelectOption[];
  disabled?: boolean;
  "aria-label"?: string;
}

// A styled dropdown over Radix Select, for a fixed small set of choices that
// needs more room than a segmented control gives it.
export function Select({ value, onValueChange, options, disabled, ...aria }: SelectProps) {
  return (
    <SelectPrimitive.Root value={value} onValueChange={onValueChange} disabled={disabled}>
      <SelectPrimitive.Trigger
        {...aria}
        className="border-border bg-surface-raised text-text focus-visible:outline-accent inline-flex items-center gap-2 rounded-sm border px-3 py-1.5 text-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 disabled:text-disabled"
      >
        <SelectPrimitive.Value />
        <SelectPrimitive.Icon>▾</SelectPrimitive.Icon>
      </SelectPrimitive.Trigger>
      <SelectPrimitive.Portal>
        <SelectPrimitive.Content className="border-border bg-surface text-text overflow-hidden rounded-md border shadow-lg">
          <SelectPrimitive.Viewport className="p-1">
            {options.map((opt) => (
              <SelectPrimitive.Item
                key={opt.value}
                value={opt.value}
                className="data-[highlighted]:bg-surface-raised data-[state=checked]:text-accent flex items-center rounded-sm px-3 py-1.5 text-sm outline-none"
              >
                <SelectPrimitive.ItemText>{opt.label}</SelectPrimitive.ItemText>
              </SelectPrimitive.Item>
            ))}
          </SelectPrimitive.Viewport>
        </SelectPrimitive.Content>
      </SelectPrimitive.Portal>
    </SelectPrimitive.Root>
  );
}

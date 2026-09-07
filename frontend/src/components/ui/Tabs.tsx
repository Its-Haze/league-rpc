import type { ReactNode } from "react";
import * as TabsPrimitive from "@radix-ui/react-tabs";

export interface TabItem {
  value: string;
  label: string;
  content: ReactNode;
}

export interface TabsProps {
  defaultValue: string;
  items: TabItem[];
}

// A styled tab strip over Radix Tabs. Not used by this ticket's own screens
// yet; the per-mode override screens (12/13) are the expected first user.
export function Tabs({ defaultValue, items }: TabsProps) {
  return (
    <TabsPrimitive.Root defaultValue={defaultValue}>
      <TabsPrimitive.List className="border-border flex gap-4 border-b">
        {items.map((item) => (
          <TabsPrimitive.Trigger
            key={item.value}
            value={item.value}
            className="text-muted data-[state=active]:text-accent data-[state=active]:border-accent -mb-px border-b-2 border-transparent px-1 py-2 text-sm font-medium"
          >
            {item.label}
          </TabsPrimitive.Trigger>
        ))}
      </TabsPrimitive.List>
      {items.map((item) => (
        <TabsPrimitive.Content key={item.value} value={item.value} className="pt-4">
          {item.content}
        </TabsPrimitive.Content>
      ))}
    </TabsPrimitive.Root>
  );
}

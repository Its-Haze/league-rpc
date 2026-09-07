import { forwardRef, type ButtonHTMLAttributes } from "react";
import { Slot } from "@radix-ui/react-slot";

type Variant = "primary" | "secondary" | "ghost" | "danger";

export interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: Variant;
  /** Render as the single child element instead of a <button> (Radix asChild pattern). */
  asChild?: boolean;
}

// No font-weight here: each variant sets its own, so the two never collide as
// same-specificity utilities whose winner depends on stylesheet order.
const base =
  "inline-flex items-center justify-center gap-2 rounded-sm px-3 py-1.5 text-sm " +
  "press disabled:pointer-events-none " +
  "focus-visible:outline focus-visible:outline-2 focus-visible:outline-accent focus-visible:outline-offset-2";

// The filled variants carry near-black text on a saturated fill, which optically
// thins at this size; semibold gives those glyphs back their mass.
const variantClass: Record<Variant, string> = {
  primary: "bg-accent text-accent-text font-semibold hover:opacity-90 disabled:opacity-60",
  secondary:
    "border border-border bg-surface-raised text-text font-medium hover:bg-border-subtle " +
    "disabled:text-disabled disabled:hover:bg-surface-raised",
  ghost:
    "bg-transparent text-text font-medium hover:bg-surface-raised " +
    "disabled:text-disabled disabled:hover:bg-transparent",
  danger: "bg-danger text-danger-text font-semibold hover:opacity-90 disabled:opacity-60",
};

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  ({ variant = "secondary", className = "", asChild = false, ...props }, ref) => {
    const Comp = asChild ? Slot : "button";
    return (
      <Comp ref={ref} className={`${base} ${variantClass[variant]} ${className}`} {...props} />
    );
  },
);
Button.displayName = "Button";

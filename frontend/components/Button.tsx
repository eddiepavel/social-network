import type { ButtonHTMLAttributes } from "react";
import clsx from "clsx";

export type ButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: "solid" | "ghost" | 'danger';
  size?: "sm" | "md";
};

export default function Button({ variant = "solid", size = "md", className, ...props }: ButtonProps) {
  return (
    <button
      className={clsx("button", variant === "ghost" && "ghost", variant === "danger" && "danger", size === "sm" && "button-sm", className)}
      {...props}
    />
  );
}

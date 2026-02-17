import type { ButtonHTMLAttributes } from "react";
import clsx from "clsx";

export type ButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: "solid" | "ghost";
  size?: "sm" | "md";
};

export default function Button({ variant = "solid", size = "md", className, ...props }: ButtonProps) {
  return (
    <button
      className={clsx("button", variant === "ghost" && "ghost", size === "sm" && "button-sm", className)}
      {...props}
    />
  );
}

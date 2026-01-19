import type { ButtonHTMLAttributes } from "react";
import clsx from "clsx";

export type ButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: "solid" | "ghost";
};

export default function Button({ variant = "solid", className, ...props }: ButtonProps) {
  return (
    <button
      className={clsx("button", variant === "ghost" && "ghost", className)}
      {...props}
    />
  );
}

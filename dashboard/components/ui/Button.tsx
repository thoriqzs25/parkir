import { ReactNode, ButtonHTMLAttributes, cloneElement, isValidElement } from "react";
import { clsx } from "clsx";

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: "primary" | "secondary" | "danger" | "ghost";
  size?: "sm" | "md" | "lg";
  asChild?: boolean;
  children: ReactNode;
}

export function Button({
  variant = "primary",
  size = "md",
  asChild,
  className,
  children,
  ...props
}: ButtonProps) {
  const classes = clsx(
    "inline-flex items-center justify-center rounded-md font-medium transition-colors focus:outline-none focus:ring-2 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed",
    {
      "bg-blue-600 text-white hover:bg-blue-700 focus:ring-blue-500": variant === "primary",
      "bg-gray-200 text-gray-900 hover:bg-gray-300 focus:ring-gray-500": variant === "secondary",
      "bg-red-600 text-white hover:bg-red-700 focus:ring-red-500": variant === "danger",
      "bg-transparent text-gray-700 hover:bg-gray-100": variant === "ghost",
      "px-3 py-1.5 text-sm": size === "sm",
      "px-4 py-2 text-sm": size === "md",
      "px-6 py-3 text-base": size === "lg",
    },
    className
  );

  if (asChild && isValidElement(children)) {
    return cloneElement(children, {
      className: clsx(classes, (children.props as { className?: string }).className),
      ...props,
    } as Record<string, unknown>);
  }

  return (
    <button className={classes} {...props}>
      {children}
    </button>
  );
}

import type { HTMLAttributes, ReactNode } from "react";
import { cn } from "@/lib/cn";

type Props = HTMLAttributes<HTMLDivElement> & {
  children: ReactNode;
  size?: "narrow" | "default" | "wide";
};

const sizeClass = {
  narrow: "max-w-3xl",
  default: "max-w-6xl",
  wide: "max-w-7xl",
};

export function Container({ children, size = "default", className, ...rest }: Props) {
  return (
    <div className={cn("mx-auto w-full px-6 sm:px-8", sizeClass[size], className)} {...rest}>
      {children}
    </div>
  );
}

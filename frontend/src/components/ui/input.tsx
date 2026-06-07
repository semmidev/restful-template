import * as React from "react"
import { cn } from "@/lib/utils"

const Input = React.forwardRef<HTMLInputElement, React.InputHTMLAttributes<HTMLInputElement>>(
  ({ className, type, ...props }, ref) => {
    return (
      <input
        type={type}
        ref={ref}
        data-slot="input"
        className={cn(
          "h-12 w-full min-w-0 rounded-lg border-3 border-black bg-white px-4 py-2.5 text-base transition-all duration-100 outline-none file:inline-flex file:h-8 file:border-0 file:bg-transparent file:text-sm file:font-medium file:text-foreground placeholder:text-neutral-500 focus-visible:border-black focus-visible:shadow-brutal focus-visible:ring-0 disabled:pointer-events-none disabled:cursor-not-allowed disabled:bg-neutral-100 disabled:opacity-50 aria-invalid:border-destructive shadow-brutal-sm text-black",
          className
        )}
        {...props}
      />
    )
  }
)
Input.displayName = "Input"

export { Input }

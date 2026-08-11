import { clsx, type ClassValue } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

export const parseStringify = <T>(value: T): T => JSON.parse(JSON.stringify(value));

export function isRedirectError(error: unknown): boolean {
  if (typeof error !== 'object' || error === null) return false;
  const digest = (error as { digest?: string }).digest;
  return typeof digest === 'string' && digest.startsWith('NEXT_REDIRECT');
}

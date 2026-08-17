import { format, isValid } from 'date-fns';
import { ptBR } from 'date-fns/locale';

export { ptBR };

/**
 * Safely parses any date representation (ISO string, date-only string, or Date instance)
 * into a valid JavaScript Date object without unintended UTC midnight shifting.
 */
export function parseDateSafe(value: string | Date | undefined | null): Date | undefined {
  if (!value) return undefined;

  if (value instanceof Date) {
    return isValid(value) ? value : undefined;
  }

  if (typeof value === 'string') {
    const trimmed = value.trim();
    if (!trimmed) return undefined;

    // Handle date-only strings (YYYY-MM-DD) to prevent JS UTC-midnight shifting to previous day 21:00
    const dateOnlyMatch = trimmed.match(/^(\d{4})-(\d{2})-(\d{2})$/);
    if (dateOnlyMatch) {
      const year = parseInt(dateOnlyMatch[1], 10);
      const month = parseInt(dateOnlyMatch[2], 10) - 1;
      const day = parseInt(dateOnlyMatch[3], 10);
      const d = new Date(year, month, day, 12, 0, 0); // Noon local avoids any DST/timezone day boundaries
      return isValid(d) ? d : undefined;
    }

    // Normalize SQL format "YYYY-MM-DD HH:mm:ss" to ISO "YYYY-MM-DDTHH:mm:ss"
    const normalized = trimmed.includes(' ') && !trimmed.includes('T')
      ? trimmed.replace(' ', 'T')
      : trimmed;

    const parsed = new Date(normalized);
    return isValid(parsed) ? parsed : undefined;
  }

  return undefined;
}

/**
 * Formats a date into Brazilian date format: "dd/MM/yyyy"
 */
export function formatDate(value: string | Date | undefined | null, fallback = '—'): string {
  const parsed = parseDateSafe(value);
  if (!parsed) return fallback;
  try {
    return format(parsed, 'dd/MM/yyyy', { locale: ptBR });
  } catch {
    return fallback;
  }
}

/**
 * Formats a date into Brazilian datetime format: "dd/MM/yyyy HH:mm"
 */
export function formatDateTime(value: string | Date | undefined | null, fallback = '—'): string {
  const parsed = parseDateSafe(value);
  if (!parsed) return fallback;
  try {
    return format(parsed, 'dd/MM/yyyy HH:mm', { locale: ptBR });
  } catch {
    return fallback;
  }
}

/**
 * Formats the time portion as "HH:mm"
 */
export function formatTime(value: string | Date | undefined | null, fallback = '00:00'): string {
  const parsed = parseDateSafe(value);
  if (!parsed) return fallback;
  try {
    return format(parsed, 'HH:mm', { locale: ptBR });
  } catch {
    return fallback;
  }
}

/**
 * Formats a date in extended text format (e.g. "17 de agosto de 2026")
 */
export function formatDateLong(value: string | Date | undefined | null, fallback = 'Selecione a data'): string {
  const parsed = parseDateSafe(value);
  if (!parsed) return fallback;
  try {
    return format(parsed, 'PPP', { locale: ptBR });
  } catch {
    return fallback;
  }
}

/**
 * Combines a Date object (with year/month/day) and a time string ("HH:mm") into a complete ISO 8601 string.
 */
export function combineDateAndTime(date: Date | undefined, timeStr: string): string {
  if (!date) return '';
  const [hours, minutes] = (timeStr || '00:00').split(':').map(Number);
  const d = new Date(date);
  d.setHours(isNaN(hours) ? 0 : hours, isNaN(minutes) ? 0 : minutes, 0, 0);
  return d.toISOString();
}

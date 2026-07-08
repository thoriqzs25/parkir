import { format, parseISO } from "date-fns";
import { toZonedTime } from "date-fns-tz";

export const JAKARTA_TZ = "Asia/Jakarta";

export function formatWIB(isoString: string, fmt = "dd/MM/yyyy HH:mm"): string {
  const date = parseISO(isoString);
  const zoned = toZonedTime(date, JAKARTA_TZ);
  return format(zoned, fmt);
}

export function formatWIBDate(isoString: string): string {
  return formatWIB(isoString, "dd/MM/yyyy");
}

export function formatWIBDateTime(isoString: string): string {
  return formatWIB(isoString, "dd/MM/yyyy HH:mm");
}

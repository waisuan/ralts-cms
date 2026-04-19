/** Machine detail page URL (lists + maintenance UI on the same page). */
export function machineDetailHref(serialNumber: string): string {
  return `/machines/${encodeURIComponent(serialNumber)}`;
}

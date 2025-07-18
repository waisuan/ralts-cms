/**
 * Converts a backend RFC3339 date string to HTML date input format (YYYY-MM-DD)
 * @param dateString - RFC3339 date string from backend (e.g., "2025-09-02T00:00:00Z")
 * @returns Date string in YYYY-MM-DD format for HTML date inputs
 */
export function backendDateToHtmlDate(dateString: string): string {
  if (!dateString) return '';
  
  try {
    const date = new Date(dateString);
    if (isNaN(date.getTime())) return '';
    
    return date.toISOString().split('T')[0]; // Returns YYYY-MM-DD
  } catch (error) {
    console.error('Error converting backend date to HTML date:', error);
    return '';
  }
}

/**
 * Converts an HTML date input value (YYYY-MM-DD) to backend RFC3339 format
 * @param htmlDate - Date string from HTML date input (e.g., "2025-09-02")
 * @returns RFC3339 date string for backend
 */
export function htmlDateToBackendDate(htmlDate: string): string {
  if (!htmlDate) return '';
  
  try {
    const date = new Date(htmlDate + 'T00:00:00Z');
    if (isNaN(date.getTime())) return '';
    
    return date.toISOString();
  } catch (error) {
    console.error('Error converting HTML date to backend date:', error);
    return '';
  }
} 
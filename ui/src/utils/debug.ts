// Debug utility for API responses

export function logApiResponse(endpoint: string, response: unknown) {
  console.log(`API Response for ${endpoint}:`, {
    response,
    hasData: response && typeof response === 'object' && 'data' in response,
    dataType: response && typeof response === 'object' && 'data' in response ? typeof (response as { data: unknown }).data : 'undefined',
    keys: response && typeof response === 'object' ? Object.keys(response as Record<string, unknown>) : [],
  });
}

export function logApiError(endpoint: string, error: unknown) {
  console.error(`API Error for ${endpoint}:`, {
    error,
    message: error && typeof error === 'object' && 'message' in error ? String((error as { message: unknown }).message) : 'Unknown error',
    status: error && typeof error === 'object' && 'status' in error ? (error as { status: unknown }).status : 'Unknown status',
    details: error && typeof error === 'object' && 'details' in error ? (error as { details: unknown }).details : 'No details',
  });
} 
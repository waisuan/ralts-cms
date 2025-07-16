// Debug utility for API responses

export function logApiResponse(endpoint: string, response: any) {
  console.log(`API Response for ${endpoint}:`, {
    response,
    hasData: 'data' in response,
    dataType: typeof response.data,
    keys: Object.keys(response),
  });
}

export function logApiError(endpoint: string, error: any) {
  console.error(`API Error for ${endpoint}:`, {
    error,
    message: error.message,
    status: error.status,
    details: error.details,
  });
} 
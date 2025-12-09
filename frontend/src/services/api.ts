import { API_BASE_URL } from '../utils/constants';

// API Error class for better error handling
export class APIError extends Error {
  status: number;

  constructor(status: number, message: string) {
    super(message);
    this.name = 'APIError';
    this.status = status;
  }
}

// Generic API request wrapper using native Fetch
export async function apiRequest<T>(
  endpoint: string,
  options: RequestInit = {}
): Promise<T> {
  const url = `${API_BASE_URL}${endpoint}`;

  const config: RequestInit = {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options.headers,
    },
    credentials: 'include', // Important: Include cookies in requests
  };

  try {
    const response = await fetch(url, config);

    // Handle different response types
    const contentType = response.headers.get('content-type');
    let payload: any;

    if (contentType && contentType.includes('application/json')) {
      payload = await response.json();
    } else {
      payload = await response.text();
    }

    // API returns envelope { data, error }, so unwrap when present
    const data =
      payload && typeof payload === 'object' && 'data' in payload
        ? (payload as { data: any }).data
        : payload;
    const errorBody =
      payload && typeof payload === 'object' && 'error' in payload
        ? (payload as { error: { message?: string } }).error
        : null;

    // If response is not ok, throw APIError
    if (!response.ok) {
      throw new APIError(
        response.status,
        errorBody?.message ||
          data ||
          `HTTP error! status: ${response.status}`
      );
    }

    return data as T;
  } catch (error) {
    // Re-throw APIError as-is
    if (error instanceof APIError) {
      throw error;
    }

    // Network errors or other issues
    if (error instanceof Error) {
      throw new APIError(0, error.message);
    }

    // Unknown error
    throw new APIError(0, 'An unknown error occurred');
  }
}

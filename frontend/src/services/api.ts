import { API_BASE_URL } from '../utils/constants';

// API Error class for better error handling
export class APIError extends Error {
  constructor(public status: number, message: string) {
    super(message);
    this.name = 'APIError';
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
    let data: any;

    if (contentType && contentType.includes('application/json')) {
      data = await response.json();
    } else {
      data = await response.text();
    }

    // If response is not ok, throw APIError
    if (!response.ok) {
      throw new APIError(
        response.status,
        data.error || data || `HTTP error! status: ${response.status}`
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

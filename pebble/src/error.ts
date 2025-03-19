export interface ErrorResponse {
  code: number
  error: string
  message?: string
}

export class AppError extends Error {
  name = 'AppError'
  timestamp: Date

  constructor(
    public readonly code: number,
    public readonly message: string,
    public readonly cause?: unknown,
  ) {
    super(message)
    this.timestamp = new Date()
  }

  public get friendlyMessage(): string {
    switch (this.code) {
      case ErrorCode.BAD_REQUEST:
        return 'Invalid request parameters'
      case ErrorCode.UNAUTHORIZED:
        return 'Please log in'
      case ErrorCode.FORBIDDEN:
        return 'Access denied'
      case ErrorCode.NOT_FOUND:
        return 'Requested resource not found'
      case ErrorCode.SERVER_ERROR:
        return 'Internal server error'
      case ErrorCode.SERVICE_UNAVAILABLE:
        return 'Service temporarily unavailable'
      default:
        return this.message || 'Operation failed, please try again'
    }
  }

  public static fromResponse(response: ErrorResponse): AppError {
    return new AppError(response.code, response.message || response.error, response)
  }

  public toJSON() {
    return {
      name: this.name,
      code: this.code,
      message: this.message,
      timestamp: this.timestamp,
      cause: this.cause,
    }
  }
}

const ErrorCode = {
  // Client error
  BAD_REQUEST: 400,
  UNAUTHORIZED: 401,
  FORBIDDEN: 403,
  NOT_FOUND: 404,

  // Server error
  SERVER_ERROR: 500,
  SERVICE_UNAVAILABLE: 503,

  // Custom error
  // ...
} as const

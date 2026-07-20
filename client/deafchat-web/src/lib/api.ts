import type { LoginRequest, TokenResponse } from "./types";

const API_BASE = "/v1";

class ApiError extends Error {
  status: number;
  body: unknown;

  constructor(status: number, message: string, body: unknown) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.body = body;
  }
}

async function request<T>(
  path: string,
  options: RequestInit = {},
): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
      ...options.headers,
    },
    ...options,
  });

  const body = await res.json();

  if (!res.ok) {
    throw new ApiError(res.status, body.error ?? res.statusText, body);
  }

  return body as T;
}

export async function login(data: LoginRequest): Promise<TokenResponse> {
  return request<TokenResponse>("/tokens/authentication", {
    method: "POST",
    body: JSON.stringify(data),
  });
}

export async function register(data: {
  name: string;
  email: string;
  password: string;
}): Promise<{ user: unknown }> {
  return request("/users", {
    method: "POST",
    body: JSON.stringify(data),
  });
}

export async function healthcheck(): Promise<{ status: string; env: string }> {
  return request("/healthcheck");
}

export { ApiError };

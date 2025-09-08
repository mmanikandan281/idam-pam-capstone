const API_BASE_URL = 'http://localhost:5000/api/v1';

class ApiError extends Error {
  constructor(message: string, public status?: number) {
    super(message);
    this.name = 'ApiError';
  }
}

const getAuthHeaders = () => {
  const token = localStorage.getItem('token');
  return {
    'Content-Type': 'application/json',
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
  };
};

const handleResponse = async (response: Response) => {
  if (!response.ok) {
    const error = await response.json().catch(() => ({ error: 'Unknown error' }));
    throw new ApiError(error.error || `HTTP ${response.status}`, response.status);
  }
  return response.json();
};

export const authApi = {
  async login(username: string, password: string, totpCode?: string) {
    const response = await fetch(`${API_BASE_URL}/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username, password, totp_code: totpCode }),
    });
    return handleResponse(response);
  },

  async register(username: string, email: string, password: string) {
    const response = await fetch(`${API_BASE_URL}/auth/register`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username, email, password }),
    });
    return handleResponse(response);
  },

  async enableTOTP() {
    const response = await fetch(`${API_BASE_URL}/totp/enable`, {
      method: 'POST',
      headers: getAuthHeaders(),
    });
    return handleResponse(response);
  },
};

export const userApi = {
  async getUsers() {
    const response = await fetch(`${API_BASE_URL}/users`, {
      headers: getAuthHeaders(),
    });
    return handleResponse(response);
  },

  async getUser(id: string) {
    const response = await fetch(`${API_BASE_URL}/users/${id}`, {
      headers: getAuthHeaders(),
    });
    return handleResponse(response);
  },

  async updateUser(id: string, updates: any) {
    const response = await fetch(`${API_BASE_URL}/users/${id}`, {
      method: 'PUT',
      headers: getAuthHeaders(),
      body: JSON.stringify(updates),
    });
    return handleResponse(response);
  },

  async assignRole(userId: string, roleId: string) {
    const response = await fetch(`${API_BASE_URL}/users/${userId}/roles`, {
      method: 'POST',
      headers: getAuthHeaders(),
      body: JSON.stringify({ role_id: roleId }),
    });
    return handleResponse(response);
  },
};

export const secretApi = {
  async getSecrets() {
    const response = await fetch(`${API_BASE_URL}/secrets`, {
      headers: getAuthHeaders(),
    });
    return handleResponse(response);
  },

  async getSecret(id: string) {
    const response = await fetch(`${API_BASE_URL}/secrets/${id}`, {
      headers: getAuthHeaders(),
    });
    return handleResponse(response);
  },

  async createSecret(data: { name: string; description: string; data: string }) {
    const response = await fetch(`${API_BASE_URL}/secrets`, {
      method: 'POST',
      headers: getAuthHeaders(),
      body: JSON.stringify(data),
    });
    return handleResponse(response);
  },

  async deleteSecret(id: string) {
    const response = await fetch(`${API_BASE_URL}/secrets/${id}`, {
      method: 'DELETE',
      headers: getAuthHeaders(),
    });
    return handleResponse(response);
  },
};

export const auditApi = {
  async getAuditLogs(limit = 100, offset = 0) {
    const response = await fetch(`${API_BASE_URL}/audit?limit=${limit}&offset=${offset}`, {
      headers: getAuthHeaders(),
    });
    return handleResponse(response);
  },
};
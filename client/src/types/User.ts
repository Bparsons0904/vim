export interface User {
  id: string;
  firstName: string;
  lastName: string;
  displayName: string;
  login: string;
  email: string;
  isAdmin: boolean;
  createdAt: string;
  updatedAt: string;
  deletedAt?: string;
}

export interface LoginRequest {
  login: string;
  password: string;
}

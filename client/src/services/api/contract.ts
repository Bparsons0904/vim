import { z } from 'zod';
import { User, LoginRequest, RegisterRequest } from '../../types/User';

export const apiContract = {
  auth: {
    login: {
      path: '/users/login',
      method: 'POST',
      body: LoginRequest,
      response: z.object({ user: User, token: z.string() }),
    },
    logout: {
      path: '/users/logout',
      method: 'POST',
      response: z.object({ message: z.string() }),
    },
    me: {
      path: '/users',
      method: 'GET',
      response: z.object({ user: User }),
    },
  },
  users: {
    get: {
      path: '/users',
      method: 'GET',
      response: z.array(User),
    },
    getById: {
      path: (id: string) => `/users/${id}`,
      method: 'GET',
      response: User,
    },
    create: {
      path: '/users',
      method: 'POST',
      body: RegisterRequest,
      response: User,
    },
  },
} as const;

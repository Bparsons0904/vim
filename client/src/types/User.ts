import { z } from 'zod';

export const User = z.object({
  id: z.string(),
  firstName: z.string(),
  lastName: z.string(),
  displayName: z.string(),
  login: z.string(),
  email: z.string().email(),
  isAdmin: z.boolean(),
  createdAt: z.string().datetime(),
  updatedAt: z.string().datetime(),
  deletedAt: z.string().datetime().optional(),
});

export type User = z.infer<typeof User>;

export const LoginRequest = z.object({
  login: z.string(),
  password: z.string(),
});

export type LoginRequest = z.infer<typeof LoginRequest>;

export const RegisterRequest = z.object({
  firstName: z.string(),
  lastName: z.string(),
  displayName: z.string(),
  login: z.string(),
  email: z.string().email(),
  password: z.string(),
});

export type RegisterRequest = z.infer<typeof RegisterRequest>;

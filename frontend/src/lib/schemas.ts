import { z } from 'zod';

export const loginSchema = z.object({
  email: z.string().email('Please enter a valid email address').min(3, 'Email is too short'),
  password: z.string().min(8, 'Password must be at least 8 characters long'),
});

export const registerSchema = z.object({
  email: z.string().email('Please enter a valid email address').min(3, 'Email is too short'),
  password: z.string().min(8, 'Password must be at least 8 characters long'),
  confirmPassword: z.string().min(8, 'Confirm password must be at least 8 characters long'),
}).refine((data) => data.password === data.confirmPassword, {
  message: 'Passwords do not match',
  path: ['confirmPassword'],
});

export const todoSchema = z.object({
  title: z.string().min(1, 'Title cannot be empty').max(200, 'Title is too long (max 200 chars)'),
  description: z.string().max(2000, 'Description is too long (max 2000 chars)').optional().or(z.literal('')),
});

export type LoginInput = z.infer<typeof loginSchema>;
export type RegisterInput = z.infer<typeof registerSchema>;
export type TodoInput = z.infer<typeof todoSchema>;

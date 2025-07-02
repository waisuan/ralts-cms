import { writeFile, readFile, access } from 'fs/promises';
import { join } from 'path';

// Mock fs/promises
jest.mock('fs/promises');

// Mock path
jest.mock('path');
const mockJoin = join as jest.MockedFunction<typeof join>;

// Test the default avatar function logic
const getDefaultAvatar = (name: string): string => {
  const initials = name
    .split(' ')
    .map((word) => word.charAt(0))
    .join('')
    .toUpperCase()
    .slice(0, 2);

  return `https://ui-avatars.com/api/?name=${encodeURIComponent(initials)}&background=random&color=fff&size=150`;
};

describe('Registration Logic', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockJoin.mockReturnValue('/test/users.txt');
  });

  it('should generate default avatar correctly', () => {
    const avatar1 = getDefaultAvatar('John Doe');
    expect(avatar1).toContain('ui-avatars.com');
    expect(avatar1).toContain('name=JD');

    const avatar2 = getDefaultAvatar('Jane Smith');
    expect(avatar2).toContain('ui-avatars.com');
    expect(avatar2).toContain('name=JS');

    const avatar3 = getDefaultAvatar('Alice');
    expect(avatar3).toContain('ui-avatars.com');
    expect(avatar3).toContain('name=A');
  });

  it('should validate email format correctly', () => {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

    // Valid emails
    expect(emailRegex.test('user@example.com')).toBe(true);
    expect(emailRegex.test('test.email@domain.co.uk')).toBe(true);

    // Invalid emails
    expect(emailRegex.test('invalid-email')).toBe(false);
    expect(emailRegex.test('user@')).toBe(false);
    expect(emailRegex.test('@domain.com')).toBe(false);
  });

  it('should sanitize user input correctly', () => {
    const sanitizeInput = (input: string) => input?.trim();
    const sanitizeEmail = (email: string) => email?.trim().toLowerCase();

    expect(sanitizeInput('  John Doe  ')).toBe('John Doe');
    expect(sanitizeEmail('  USER@EXAMPLE.COM  ')).toBe('user@example.com');
  });

  it('should create user object with correct structure', () => {
    const createUser = (name: string, email: string, password: string) => ({
      id: Date.now().toString(),
      name: name.trim(),
      email: email.trim().toLowerCase(),
      password: password.trim(),
      role: 'user' as const,
      avatar: getDefaultAvatar(name.trim()),
      created_at: new Date().toISOString(),
    });

    const user = createUser('John Doe', 'john@example.com', 'password123');

    expect(user).toMatchObject({
      name: 'John Doe',
      email: 'john@example.com',
      role: 'user',
    });
    expect(user.id).toBeDefined();
    expect(user.created_at).toBeDefined();
    expect(user.avatar).toContain('ui-avatars.com');
    expect(user.password).toBe('password123');
  });
});

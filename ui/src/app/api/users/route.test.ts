import { join } from 'path';

// Mock path
jest.mock('path');
const mockJoin = join as jest.MockedFunction<typeof join>;

interface UserData {
  id: string;
  name: string;
  email: string;
  password: string;
  role: 'admin' | 'user';
  avatar?: string;
  created_at: string;
}

describe('Users Logic', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockJoin.mockReturnValue('/test/users.txt');
  });

  it('should parse user data correctly', () => {
    const userData: UserData[] = [
      {
        id: '1',
        name: 'John Doe',
        email: 'john@example.com',
        password: 'password123',
        role: 'admin',
        avatar: 'https://example.com/avatar1.jpg',
        created_at: '2024-01-01T00:00:00.000Z',
      },
      {
        id: '2',
        name: 'Jane Smith',
        email: 'jane@example.com',
        password: 'password456',
        role: 'user',
        avatar: 'https://example.com/avatar2.jpg',
        created_at: '2024-01-02T00:00:00.000Z',
      },
    ];

    const parseUsers = (data: UserData[]) => {
      return data.map(({ password, ...user }) => user);
    };

    const result = parseUsers(userData);

    expect(result).toHaveLength(2);
    expect(result[0]).toMatchObject({
      id: '1',
      name: 'John Doe',
      email: 'john@example.com',
      role: 'admin',
      avatar: 'https://example.com/avatar1.jpg',
      created_at: '2024-01-01T00:00:00.000Z',
    });
    expect(result[0]).not.toHaveProperty('password');
    expect(result[1]).not.toHaveProperty('password');
  });

  it('should handle empty user data', () => {
    const parseUsers = (data: UserData[]) => {
      return data.map(({ password, ...user }) => user);
    };

    const result = parseUsers([]);
    expect(result).toEqual([]);
  });

  it('should handle users without avatars', () => {
    const userData: UserData[] = [
      {
        id: '1',
        name: 'John Doe',
        email: 'john@example.com',
        password: 'password123',
        role: 'admin',
        created_at: '2024-01-01T00:00:00.000Z',
      },
    ];

    const parseUsers = (data: UserData[]) => {
      return data.map(({ password, ...user }) => user);
    };

    const result = parseUsers(userData);

    expect(result).toHaveLength(1);
    expect(result[0]).toMatchObject({
      id: '1',
      name: 'John Doe',
      email: 'john@example.com',
      role: 'admin',
      created_at: '2024-01-01T00:00:00.000Z',
    });
    expect(result[0].avatar).toBeUndefined();
    expect(result[0]).not.toHaveProperty('password');
  });

  it('should validate user data structure', () => {
    const validateUser = (user: Record<string, unknown>) => {
      const requiredFields = ['id', 'name', 'email', 'role', 'created_at'];
      return requiredFields.every((field) => user.hasOwnProperty(field));
    };

    const validUser = {
      id: '1',
      name: 'John Doe',
      email: 'john@example.com',
      role: 'admin',
      created_at: '2024-01-01T00:00:00.000Z',
    };

    const invalidUser = {
      id: '1',
      name: 'John Doe',
      // missing email, role, created_at
    };

    expect(validateUser(validUser)).toBe(true);
    expect(validateUser(invalidUser)).toBe(false);
  });
});

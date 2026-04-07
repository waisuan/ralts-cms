/** @jest-environment node */
import { mkdirSync, mkdtempSync, rmSync, writeFileSync } from 'fs';
import { join } from 'path';
import { tmpdir } from 'os';

describe('GET /api/users (route handler)', () => {
  let tmp: string;
  let cwdSpy: jest.SpiedFunction<typeof process.cwd>;

  beforeEach(() => {
    tmp = mkdtempSync(join(tmpdir(), 'ralts-api-users-'));
    mkdirSync(join(tmp, 'src', 'data'), { recursive: true });
    cwdSpy = jest.spyOn(process, 'cwd').mockReturnValue(tmp);
    jest.resetModules();
  });

  afterEach(() => {
    cwdSpy.mockRestore();
    jest.resetModules();
    rmSync(tmp, { recursive: true, force: true });
  });

  it('returns empty users when file is missing', async () => {
    const { GET } = await import('./route');
    const res = await GET();
    expect(res.status).toBe(200);
    const body = await res.json();
    expect(body).toEqual({ users: [] });
  });

  it('returns empty users when file is blank', async () => {
    writeFileSync(join(tmp, 'src', 'data', 'users.txt'), '   \n  ');
    const { GET } = await import('./route');
    const res = await GET();
    expect(res.status).toBe(200);
    expect(await res.json()).toEqual({ users: [] });
  });

  it('returns users from JSON file (passwords included as in handler)', async () => {
    const users = [
      {
        id: '1',
        name: 'Alice',
        email: 'a@example.com',
        password: 'secret',
        role: 'user' as const,
        created_at: '2024-01-01T00:00:00.000Z',
      },
    ];
    writeFileSync(join(tmp, 'src', 'data', 'users.txt'), JSON.stringify(users));
    const { GET } = await import('./route');
    const res = await GET();
    expect(res.status).toBe(200);
    const body = await res.json();
    expect(body.users).toEqual(users);
  });

  it('returns 500 when file contains invalid JSON', async () => {
    const err = jest.spyOn(console, 'error').mockImplementation(() => {});
    writeFileSync(join(tmp, 'src', 'data', 'users.txt'), 'not-json');
    const { GET } = await import('./route');
    const res = await GET();
    expect(res.status).toBe(500);
    expect(await res.json()).toEqual({ error: 'Failed to read users' });
    err.mockRestore();
  });
});

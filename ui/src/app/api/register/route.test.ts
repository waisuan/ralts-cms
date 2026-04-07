/** @jest-environment node */
import { mkdtempSync, readFileSync, rmSync, writeFileSync } from 'fs';
import { join } from 'path';
import { tmpdir } from 'os';
import { NextRequest } from 'next/server';

function jsonRequest(body: unknown) {
  return new NextRequest('http://localhost/api/register', {
    method: 'POST',
    headers: { 'content-type': 'application/json' },
    body: JSON.stringify(body),
  });
}

describe('POST /api/register (route handler)', () => {
  let tmp: string;
  let cwdSpy: jest.SpiedFunction<typeof process.cwd>;
  let nowSpy: jest.SpiedFunction<typeof Date.now>;

  beforeEach(() => {
    tmp = mkdtempSync(join(tmpdir(), 'ralts-api-register-'));
    cwdSpy = jest.spyOn(process, 'cwd').mockReturnValue(tmp);
    nowSpy = jest.spyOn(Date, 'now').mockReturnValue(1_700_000_000_000);
    jest.resetModules();
  });

  afterEach(() => {
    cwdSpy.mockRestore();
    nowSpy.mockRestore();
    jest.resetModules();
    rmSync(tmp, { recursive: true, force: true });
  });

  it('returns 400 when fields are missing', async () => {
    const { POST } = await import('./route');
    const res = await POST(jsonRequest({ name: '', email: 'a@b.com', password: 'longenough' }));
    expect(res.status).toBe(400);
    const body = await res.json();
    expect(body.error).toMatch(/required/i);
  });

  it('returns 400 for invalid email', async () => {
    const { POST } = await import('./route');
    const res = await POST(jsonRequest({ name: 'Bob', email: 'not-an-email', password: 'longenough' }));
    expect(res.status).toBe(400);
    expect(await res.json()).toMatchObject({ error: expect.stringMatching(/valid email/i) });
  });

  it('returns 400 when password is too short', async () => {
    const { POST } = await import('./route');
    const res = await POST(jsonRequest({ name: 'Bob', email: 'bob@example.com', password: '12345' }));
    expect(res.status).toBe(400);
    expect(await res.json()).toMatchObject({ error: expect.stringMatching(/6 characters/i) });
  });

  it('returns 201, writes users.txt, omits password in response', async () => {
    const { POST } = await import('./route');
    const res = await POST(
      jsonRequest({ name: '  Bob Smith  ', email: '  BOB@EXAMPLE.COM  ', password: '  secret12  ' })
    );
    expect(res.status).toBe(201);
    const body = await res.json();
    expect(body.message).toMatch(/success/i);
    expect(body.user).toMatchObject({
      id: '1700000000000',
      name: 'Bob Smith',
      email: 'bob@example.com',
      role: 'user',
    });
    expect(body.user).not.toHaveProperty('password');
    expect(body.user.avatar).toContain('ui-avatars.com');

    const stored = JSON.parse(readFileSync(join(tmp, 'users.txt'), 'utf-8')) as Array<{ password?: string }>;
    expect(stored).toHaveLength(1);
    expect(stored[0].password).toBe('secret12');
  });

  it('returns 409 when email already exists', async () => {
    const existing = [
      {
        id: 'x',
        name: 'Other',
        email: 'bob@example.com',
        password: 'p',
        role: 'user' as const,
        created_at: '2024-01-01T00:00:00.000Z',
      },
    ];
    writeFileSync(join(tmp, 'users.txt'), JSON.stringify(existing));

    const { POST } = await import('./route');
    const res = await POST(jsonRequest({ name: 'Bob', email: 'bob@example.com', password: 'longenough' }));
    expect(res.status).toBe(409);
    expect(await res.json()).toMatchObject({ error: expect.stringMatching(/email/i) });
  });

  it('returns 409 when name already exists', async () => {
    const existing = [
      {
        id: 'x',
        name: 'Bob',
        email: 'other@example.com',
        password: 'p',
        role: 'user' as const,
        created_at: '2024-01-01T00:00:00.000Z',
      },
    ];
    writeFileSync(join(tmp, 'users.txt'), JSON.stringify(existing));

    const { POST } = await import('./route');
    const res = await POST(jsonRequest({ name: 'bob', email: 'new@example.com', password: 'longenough' }));
    expect(res.status).toBe(409);
    expect(await res.json()).toMatchObject({ error: expect.stringMatching(/name/i) });
  });
});

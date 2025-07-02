import { NextResponse } from 'next/server';
import { readFile, access } from 'fs/promises';
import { join } from 'path';

const USERS_FILE = join(process.cwd(), 'users.txt');

interface UserData {
  id: string;
  name: string;
  email: string;
  password: string;
  role: 'admin' | 'user';
  created_at: string;
}

export async function GET() {
  try {
    // Check if users file exists
    try {
      await access(USERS_FILE);
    } catch {
      // File doesn't exist, return empty array
      return NextResponse.json({ users: [] });
    }

    // Read users from file
    const fileContent = await readFile(USERS_FILE, 'utf-8');
    if (!fileContent.trim()) {
      return NextResponse.json({ users: [] });
    }

    const users: UserData[] = JSON.parse(fileContent);

    // Return users with passwords for authentication (this is internal use)
    return NextResponse.json({ users });
  } catch (error) {
    console.error('Error reading users file:', error);
    return NextResponse.json({ error: 'Failed to read users' }, { status: 500 });
  }
}

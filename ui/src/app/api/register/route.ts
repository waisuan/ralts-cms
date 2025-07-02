import { NextRequest, NextResponse } from 'next/server';
import { writeFile, readFile, access } from 'fs/promises';
import { join } from 'path';

const USERS_FILE = join(process.cwd(), 'users.txt');

interface UserData {
  id: string;
  name: string;
  email: string;
  password: string;
  role: 'user';
  avatar?: string;
  created_at: string;
}

// Default avatar for users without one
const getDefaultAvatar = (name: string): string => {
  // Generate initials from name
  const initials = name
    .split(' ')
    .map((word) => word.charAt(0))
    .join('')
    .toUpperCase()
    .slice(0, 2);

  // Use a placeholder service that generates avatars with initials
  return `https://ui-avatars.com/api/?name=${encodeURIComponent(initials)}&background=random&color=fff&size=150`;
};

export async function POST(request: NextRequest) {
  try {
    const { name, email, password } = await request.json();

    // Sanitize inputs
    const sanitizedName = name?.trim();
    const sanitizedEmail = email?.trim().toLowerCase();
    const sanitizedPassword = password?.trim();

    // Validation
    if (!sanitizedName || !sanitizedEmail || !sanitizedPassword) {
      return NextResponse.json(
        { error: 'Name, email, and password are required' },
        { status: 400 }
      );
    }

    // Email format validation
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailRegex.test(sanitizedEmail)) {
      return NextResponse.json({ error: 'Please enter a valid email address' }, { status: 400 });
    }

    if (sanitizedPassword.length < 6) {
      return NextResponse.json(
        { error: 'Password must be at least 6 characters long' },
        { status: 400 }
      );
    }

    // Check if user already exists
    let existingUsers: UserData[] = [];
    try {
      await access(USERS_FILE);
      const fileContent = await readFile(USERS_FILE, 'utf-8');
      if (fileContent.trim()) {
        existingUsers = JSON.parse(fileContent);
      }
    } catch {
      // File doesn't exist, which is fine for new installations
    }

    // Check if email already exists (case-insensitive)
    if (existingUsers.some((user) => user.email.toLowerCase() === sanitizedEmail)) {
      return NextResponse.json({ error: 'User with this email already exists' }, { status: 409 });
    }

    // Optional: Check if name already exists (case-insensitive)
    if (existingUsers.some((user) => user.name.toLowerCase() === sanitizedName.toLowerCase())) {
      return NextResponse.json({ error: 'User with this name already exists' }, { status: 409 });
    }

    // Create new user with default avatar
    const newUser: UserData = {
      id: Date.now().toString(),
      name: sanitizedName,
      email: sanitizedEmail,
      password: sanitizedPassword, // In a real app, this should be hashed
      role: 'user',
      avatar: getDefaultAvatar(sanitizedName),
      created_at: new Date().toISOString(),
    };

    // Add to existing users
    existingUsers.push(newUser);

    // Save to file
    await writeFile(USERS_FILE, JSON.stringify(existingUsers, null, 2));

    // Return success (don't return the password)
    const { password: _pw, ...userWithoutPassword } = newUser; // eslint-disable-line @typescript-eslint/no-unused-vars
    return NextResponse.json(
      {
        message: 'User registered successfully',
        user: userWithoutPassword,
      },
      { status: 201 }
    );
  } catch (error) {
    console.error('Registration error:', error);
    return NextResponse.json({ error: 'Internal server error' }, { status: 500 });
  }
}

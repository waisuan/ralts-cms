import { NextRequest, NextResponse } from 'next/server';
import { writeFile, readFile, access } from 'fs/promises';
import { join } from 'path';
import { Machine } from '../../../types/machine';

const MACHINES_FILE = join(process.cwd(), 'src/data/machines.txt');

// Helper to read all machines
async function readMachines(): Promise<Machine[]> {
  try {
    await access(MACHINES_FILE);
    const fileContent = await readFile(MACHINES_FILE, 'utf-8');
    return fileContent.trim() ? JSON.parse(fileContent) : [];
  } catch {
    return [];
  }
}

// Helper to write all machines
async function writeMachines(machines: Machine[]): Promise<void> {
  await writeFile(MACHINES_FILE, JSON.stringify(machines, null, 2));
}

export async function GET() {
  try {
    const machines = await readMachines();
    return NextResponse.json(machines);
  } catch (error) {
    return NextResponse.json({ error: 'Failed to read machines' }, { status: 500 });
  }
}

export async function POST(request: NextRequest) {
  try {
    const data = await request.json();
    // Validate required fields
    const requiredFields = [
      'serial_number', 'customer', 'state', 'account_type', 'model', 'status', 'brand', 'district',
      'person_in_charge', 'reported_by', 'additional_notes', 'attachment', 'ppm_status', 'tnc_date', 'ppm_date'
    ];
    for (const field of requiredFields) {
      if (!data[field] && data[field] !== '') {
        return NextResponse.json({ error: `Missing field: ${field}` }, { status: 400 });
      }
    }
    const now = new Date().toISOString();
    const newMachine: Machine = {
      ...data,
      created_at: now,
      updated_at: now,
    };
    const machines = await readMachines();
    // Ensure unique serial_number
    if (machines.some(m => m.serial_number === newMachine.serial_number)) {
      return NextResponse.json({ error: 'Machine with this serial_number already exists' }, { status: 409 });
    }
    machines.push(newMachine);
    await writeMachines(machines);
    return NextResponse.json(newMachine, { status: 201 });
  } catch (error) {
    return NextResponse.json({ error: 'Failed to create machine' }, { status: 500 });
  }
} 
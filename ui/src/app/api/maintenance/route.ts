import { NextRequest, NextResponse } from 'next/server';
import { writeFile, readFile, access } from 'fs/promises';
import { join } from 'path';
import { Maintenance } from '../../../types/maintenance';

const MAINTENANCE_FILE = join(process.cwd(), 'src/data/maintenance.txt');

async function readMaintenance(): Promise<Maintenance[]> {
  try {
    await access(MAINTENANCE_FILE);
    const fileContent = await readFile(MAINTENANCE_FILE, 'utf-8');
    return fileContent.trim() ? JSON.parse(fileContent) : [];
  } catch {
    return [];
  }
}

async function writeMaintenance(records: Maintenance[]): Promise<void> {
  await writeFile(MAINTENANCE_FILE, JSON.stringify(records, null, 2));
}

export async function GET() {
  try {
    const records = await readMaintenance();
    return NextResponse.json(records);
  } catch (error) {
    return NextResponse.json({ error: 'Failed to read maintenance records' }, { status: 500 });
  }
}

export async function POST(request: NextRequest) {
  try {
    const data = await request.json();
    // Validate required fields
    const requiredFields = [
      'machine_serial_number', 'work_order_number', 'work_order_date', 'action_taken', 'reported_by', 'worker_order_type', 'attachment'
    ];
    for (const field of requiredFields) {
      if (!data[field] && data[field] !== '') {
        return NextResponse.json({ error: `Missing field: ${field}` }, { status: 400 });
      }
    }
    const now = new Date().toISOString();
    const newRecord: Maintenance = {
      ...data,
      created_at: now,
      updated_at: now,
    };
    const records = await readMaintenance();
    // Ensure unique work_order_number
    if (records.some(r => r.work_order_number === newRecord.work_order_number)) {
      return NextResponse.json({ error: 'Maintenance record with this work_order_number already exists' }, { status: 409 });
    }
    records.push(newRecord);
    await writeMaintenance(records);
    return NextResponse.json(newRecord, { status: 201 });
  } catch (error) {
    return NextResponse.json({ error: 'Failed to create maintenance record' }, { status: 500 });
  }
} 
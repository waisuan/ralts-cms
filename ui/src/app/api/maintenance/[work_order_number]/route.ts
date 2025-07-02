import { NextRequest, NextResponse } from 'next/server';
import { writeFile, readFile, access } from 'fs/promises';
import { join } from 'path';
import { Maintenance } from '../../../../types/maintenance';

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

export async function GET(
  _request: NextRequest,
  { params }: { params: { work_order_number: string } }
) {
  try {
    const records = await readMaintenance();
    const record = records.find(r => r.work_order_number === params.work_order_number);
    if (!record) {
      return NextResponse.json({ error: 'Maintenance record not found' }, { status: 404 });
    }
    return NextResponse.json(record);
  } catch (error) {
    return NextResponse.json({ error: 'Failed to read maintenance record' }, { status: 500 });
  }
}

export async function PUT(
  request: NextRequest,
  { params }: { params: { work_order_number: string } }
) {
  try {
    const data = await request.json();
    const records = await readMaintenance();
    const idx = records.findIndex(r => r.work_order_number === params.work_order_number);
    if (idx === -1) {
      return NextResponse.json({ error: 'Maintenance record not found' }, { status: 404 });
    }
    const updatedRecord: Maintenance = {
      ...records[idx],
      ...data,
      updated_at: new Date().toISOString(),
    };
    records[idx] = updatedRecord;
    await writeMaintenance(records);
    return NextResponse.json(updatedRecord);
  } catch (error) {
    return NextResponse.json({ error: 'Failed to update maintenance record' }, { status: 500 });
  }
}

export async function DELETE(
  _request: NextRequest,
  { params }: { params: { work_order_number: string } }
) {
  try {
    const records = await readMaintenance();
    const idx = records.findIndex(r => r.work_order_number === params.work_order_number);
    if (idx === -1) {
      return NextResponse.json({ error: 'Maintenance record not found' }, { status: 404 });
    }
    const deleted = records.splice(idx, 1)[0];
    await writeMaintenance(records);
    return NextResponse.json(deleted);
  } catch (error) {
    return NextResponse.json({ error: 'Failed to delete maintenance record' }, { status: 500 });
  }
} 
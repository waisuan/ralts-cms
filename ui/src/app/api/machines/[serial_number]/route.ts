import { NextRequest, NextResponse } from 'next/server';
import { writeFile, readFile, access } from 'fs/promises';
import { join } from 'path';
import { Machine } from '../../../../types/machine';

const MACHINES_FILE = join(process.cwd(), 'src/data/machines.txt');

async function readMachines(): Promise<Machine[]> {
  try {
    await access(MACHINES_FILE);
    const fileContent = await readFile(MACHINES_FILE, 'utf-8');
    return fileContent.trim() ? JSON.parse(fileContent) : [];
  } catch {
    return [];
  }
}

async function writeMachines(machines: Machine[]): Promise<void> {
  await writeFile(MACHINES_FILE, JSON.stringify(machines, null, 2));
}

export async function GET(
  _request: NextRequest,
  { params }: { params: { serial_number: string } }
) {
  try {
    const machines = await readMachines();
    const machine = machines.find(m => m.serial_number === params.serial_number);
    if (!machine) {
      return NextResponse.json({ error: 'Machine not found' }, { status: 404 });
    }
    return NextResponse.json(machine);
  } catch (error) {
    return NextResponse.json({ error: 'Failed to read machine' }, { status: 500 });
  }
}

export async function PUT(
  request: NextRequest,
  { params }: { params: { serial_number: string } }
) {
  try {
    const data = await request.json();
    const machines = await readMachines();
    const idx = machines.findIndex(m => m.serial_number === params.serial_number);
    if (idx === -1) {
      return NextResponse.json({ error: 'Machine not found' }, { status: 404 });
    }
    const updatedMachine: Machine = {
      ...machines[idx],
      ...data,
      updated_at: new Date().toISOString(),
    };
    machines[idx] = updatedMachine;
    await writeMachines(machines);
    return NextResponse.json(updatedMachine);
  } catch (error) {
    return NextResponse.json({ error: 'Failed to update machine' }, { status: 500 });
  }
}

export async function DELETE(
  _request: NextRequest,
  { params }: { params: { serial_number: string } }
) {
  try {
    const machines = await readMachines();
    const idx = machines.findIndex(m => m.serial_number === params.serial_number);
    if (idx === -1) {
      return NextResponse.json({ error: 'Machine not found' }, { status: 404 });
    }
    const deleted = machines.splice(idx, 1)[0];
    await writeMachines(machines);
    return NextResponse.json(deleted);
  } catch (error) {
    return NextResponse.json({ error: 'Failed to delete machine' }, { status: 500 });
  }
} 
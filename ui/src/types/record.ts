export interface Record {
  id: number;
  title: string;
  description: string;
  status: 'active' | 'inactive' | 'pending';
  createdAt: string;
  updatedAt: string;
} 
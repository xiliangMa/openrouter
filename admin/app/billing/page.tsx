'use client';

import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import ProtectedRoute from '@/components/protected-route';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Plus, Search, DollarSign, CreditCard, History, TrendingUp, Download } from 'lucide-react';
import api from '@/lib/api';

interface BalanceInfo {
  balance: number;
  is_overdue: boolean;
}

interface Payment {
  id: string;
  amount: number;
  currency: string;
  status: string;
  payment_method: string;
  description: string;
  created_at: string;
  completed_at: string | null;
}

interface BillingRecord {
  id: string;
  user_id: string;
  model_id: string;
  api_key_id: string | null;
  input_tokens: number;
  output_tokens: number;
  cost: number;
  currency: string;
  timestamp: string;
  model: {
    name: string;
    provider_name: string;
  };
  user: {
    email: string;
    username: string;
  };
}

export default function BillingPage() {
  const [search, setSearch] = useState('');
  const [page, setPage] = useState(1);
  const limit = 20;

  const { data: balanceData, isLoading: balanceLoading } = useQuery({
    queryKey: ['billing-balance'],
    queryFn: async () => {
      const response = await api.get('/billing/balance');
      return response.data.data;
    },
  });

  const { data: paymentsData, isLoading: paymentsLoading } = useQuery({
    queryKey: ['billing-payments', page],
    queryFn: async () => {
      const response = await api.get('/billing/payments', {
        params: { page, limit },
      });
      return response.data.data;
    },
  });

  const { data: recordsData, isLoading: recordsLoading } = useQuery({
    queryKey: ['billing-records', page],
    queryFn: async () => {
      const response = await api.get('/billing/records', {
        params: { page, limit },
      });
      return response.data.data;
    },
  });

  const balance: BalanceInfo = balanceData || { balance: 0, is_overdue: false };
  const payments: Payment[] = paymentsData?.payments || [];
  const records: BillingRecord[] = recordsData?.records || [];

  const totalRevenue = records.reduce((sum, record) => sum + record.cost, 0);
  const avgTransaction = records.length > 0 ? totalRevenue / records.length : 0;

  return (
    <ProtectedRoute requireAdmin={true}>
      <div className="p-6 space-y-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Billing Management</h1>
            <p className="text-muted-foreground">
              Monitor revenue, payments, and billing records
            </p>
          </div>
          <div className="flex items-center gap-2">
            <Button variant="outline">
              <Download className="mr-2 h-4 w-4" />
              Export Data
            </Button>
            <Dialog>
              <DialogTrigger asChild>
                <Button>
                  <Plus className="mr-2 h-4 w-4" />
                  Add Credit
                </Button>
              </DialogTrigger>
              <DialogContent className="sm:max-w-[400px]">
                <DialogHeader>
                  <DialogTitle>Add Credit to User</DialogTitle>
                  <DialogDescription>
                    Add credit to a user&apos;s account manually.
                  </DialogDescription>
                </DialogHeader>
                <div className="grid gap-4 py-4">
                  <div className="grid gap-2">
                    <Label htmlFor="user-email">User Email</Label>
                    <Input
                      id="user-email"
                      placeholder="user@example.com"
                    />
                  </div>
                  <div className="grid gap-2">
                    <Label htmlFor="amount">Amount (USD)</Label>
                    <Input
                      id="amount"
                      type="number"
                      step="0.01"
                      placeholder="100.00"
                    />
                  </div>
                  <div className="grid gap-2">
                    <Label htmlFor="reason">Reason</Label>
                    <Input
                      id="reason"
                      placeholder="Manual credit addition"
                    />
                  </div>
                </div>
                <DialogFooter>
                  <Button type="button" variant="outline">Cancel</Button>
                  <Button type="submit">Add Credit</Button>
                </DialogFooter>
              </DialogContent>
            </Dialog>
          </div>
        </div>

        {/* Stats Cards */}
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-4">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Account Balance</CardTitle>
              <DollarSign className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                ${balance.balance?.toFixed(2) || '0.00'}
              </div>
              <p className="text-xs text-muted-foreground">
                Status: ${balance.is_overdue ? 'Overdue' : 'Good standing'}
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Total Revenue</CardTitle>
              <TrendingUp className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">${totalRevenue.toFixed(2)}</div>
              <p className="text-xs text-muted-foreground">
                From {records.length} transactions
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Recent Payments</CardTitle>
              <CreditCard className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{payments.length}</div>
              <p className="text-xs text-muted-foreground">
                {payments.filter(p => p.status === 'completed').length} completed
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Avg. Transaction</CardTitle>
              <History className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">${avgTransaction.toFixed(4)}</div>
              <p className="text-xs text-muted-foreground">
                Per API usage record
              </p>
            </CardContent>
          </Card>
        </div>

        <div className="grid gap-6 lg:grid-cols-2">
          {/* Recent Payments */}
          <Card>
            <CardHeader>
              <CardTitle>Recent Payments</CardTitle>
              <CardDescription>
                Latest payment transactions
              </CardDescription>
            </CardHeader>
            <CardContent>
              {paymentsLoading ? (
                <div className="text-center py-8">Loading payments...</div>
              ) : payments.length === 0 ? (
                <div className="text-center py-8 text-muted-foreground">No payment history</div>
              ) : (
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Amount</TableHead>
                      <TableHead>Method</TableHead>
                      <TableHead>Status</TableHead>
                      <TableHead>Date</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {payments.slice(0, 5).map((payment) => (
                      <TableRow key={payment.id}>
                        <TableCell className="font-medium">
                          ${payment.amount.toFixed(2)}
                        </TableCell>
                        <TableCell>{payment.payment_method}</TableCell>
                        <TableCell>
                          <Badge variant={
                            payment.status === 'completed' ? 'default' :
                            payment.status === 'pending' ? 'secondary' : 'destructive'
                          }>
                            {payment.status}
                          </Badge>
                        </TableCell>
                        <TableCell>
                          {new Date(payment.created_at).toLocaleDateString()}
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              )}
              {payments.length > 5 && (
                <div className="mt-4 text-center">
                  <Button variant="outline" size="sm">
                    View All Payments
                  </Button>
                </div>
              )}
            </CardContent>
          </Card>

          {/* Billing Records */}
          <Card>
            <CardHeader>
              <CardTitle>Recent Billing Records</CardTitle>
              <CardDescription>
                API usage and cost records
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="flex items-center gap-4 mb-6">
                <div className="relative flex-1">
                  <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground h-4 w-4" />
                  <Input
                    placeholder="Search records..."
                    value={search}
                    onChange={(e) => setSearch(e.target.value)}
                    className="pl-10"
                  />
                </div>
              </div>

              {recordsLoading ? (
                <div className="text-center py-8">Loading records...</div>
              ) : records.length === 0 ? (
                <div className="text-center py-8 text-muted-foreground">No billing records</div>
              ) : (
                <div className="space-y-4">
                  {records
                    .filter(record =>
                      record.user.email.toLowerCase().includes(search.toLowerCase()) ||
                      record.model.name.toLowerCase().includes(search.toLowerCase())
                    )
                    .slice(0, 5)
                    .map((record) => (
                      <div key={record.id} className="flex items-center justify-between p-3 border rounded-lg">
                        <div>
                          <div className="font-medium">{record.user.email}</div>
                          <div className="text-sm text-muted-foreground">
                            {record.model.name} • {record.input_tokens + record.output_tokens} tokens
                          </div>
                        </div>
                        <div className="text-right">
                          <div className="font-medium">${record.cost.toFixed(6)}</div>
                          <div className="text-sm text-muted-foreground">
                            {new Date(record.timestamp).toLocaleDateString()}
                          </div>
                        </div>
                      </div>
                    ))}
                </div>
              )}
              {records.length > 5 && (
                <div className="mt-4 text-center">
                  <Button variant="outline" size="sm">
                    View All Records
                  </Button>
                </div>
              )}
            </CardContent>
          </Card>
        </div>
      </div>
    </ProtectedRoute>
  );
}
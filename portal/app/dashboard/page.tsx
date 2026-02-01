'use client';

import ProtectedRoute from '@/components/protected-route';
import { useAuth } from '@/lib/auth';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { BarChart3, CreditCard, Key, Zap, Activity, DollarSign, Database, Clock, Users, Settings, Bell, ChevronDown } from 'lucide-react';
import { useQuery } from '@tanstack/react-query';
import api from '@/lib/api';

export default function DashboardPage() {
  const { user, logout } = useAuth();

  const { data: balanceData, isLoading: balanceLoading } = useQuery({
    queryKey: ['user-balance'],
    queryFn: async () => {
      const response = await api.get('/user/balance');
      return response.data.data;
    },
    enabled: !!user,
  });

  const { data: usageData, isLoading: usageLoading } = useQuery({
    queryKey: ['user-usage'],
    queryFn: async () => {
      const response = await api.get('/user/usage');
      return response.data.data;
    },
    enabled: !!user,
  });

  const { data: apiKeysData, isLoading: apiKeysLoading } = useQuery({
    queryKey: ['user-api-keys'],
    queryFn: async () => {
      const response = await api.get('/user/api-keys');
      return response.data.data;
    },
    enabled: !!user,
  });

  return (
    <ProtectedRoute>
      <div className="min-h-screen bg-gray-50">
        {/* Navigation */}
        <nav className="sticky top-0 z-50 border-b bg-white shadow-sm">
          <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
            <div className="flex h-16 justify-between">
              <div className="flex">
                <div className="flex flex-shrink-0 items-center">
                  <Zap className="h-8 w-8 text-primary" />
                  <span className="ml-2 text-xl font-bold text-gray-900">MassRouter</span>
                </div>
                <div className="hidden sm:ml-6 sm:flex sm:space-x-8">
                  <a href="/dashboard" className="inline-flex items-center border-b-2 border-primary px-1 pt-1 text-sm font-medium text-gray-900">
                    Dashboard
                  </a>
                  <a href="/models" className="inline-flex items-center border-b-2 border-transparent px-1 pt-1 text-sm font-medium text-gray-500 hover:border-gray-300 hover:text-gray-700">
                    Models
                  </a>
                  <a href="/api-keys" className="inline-flex items-center border-b-2 border-transparent px-1 pt-1 text-sm font-medium text-gray-500 hover:border-gray-300 hover:text-gray-700">
                    API Keys
                  </a>
                  <a href="/billing" className="inline-flex items-center border-b-2 border-transparent px-1 pt-1 text-sm font-medium text-gray-500 hover:border-gray-300 hover:text-gray-700">
                    Billing
                  </a>
                </div>
              </div>
              <div className="hidden sm:ml-6 sm:flex sm:items-center">
                <button className="rounded-full bg-white p-1 text-gray-400 hover:text-gray-500 focus:outline-none focus:ring-2 focus:ring-primary focus:ring-offset-2">
                  <Bell className="h-6 w-6" />
                </button>
                <div className="relative ml-3">
                  <div className="flex items-center space-x-3">
                    <div className="text-right">
                      <p className="text-sm font-medium text-gray-900">{user?.username}</p>
                      <p className="text-xs text-gray-500">{user?.email}</p>
                    </div>
                    <div className="h-8 w-8 rounded-full bg-gradient-primary flex items-center justify-center text-white font-medium">
                      {user?.username?.charAt(0).toUpperCase()}
                    </div>
                    <ChevronDown className="h-5 w-5 text-gray-400" />
                  </div>
                </div>
              </div>
            </div>
          </div>
        </nav>

        <main className="py-8">
          <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
            {/* Header */}
            <div className="mb-8">
              <h1 className="text-3xl font-bold text-gray-900">Welcome back, {user?.username}!</h1>
              <p className="mt-2 text-gray-600">
                Here&apos;s what&apos;s happening with your MassRouter account today.
              </p>
            </div>

            {/* Stats Grid */}
            <div className="mb-8 grid gap-6 sm:grid-cols-2 lg:grid-cols-4">
              <Card className="card-hover">
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                  <CardTitle className="text-sm font-medium text-gray-600">Account Balance</CardTitle>
                  <CreditCard className="h-4 w-4 text-primary" />
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold text-gray-900">${balanceLoading ? '...' : (balanceData?.balance || 0).toFixed(2)}</div>
                  <p className="text-xs text-gray-500">Available credit</p>
                </CardContent>
              </Card>

              <Card className="card-hover">
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                  <CardTitle className="text-sm font-medium text-gray-600">API Keys</CardTitle>
                  <Key className="h-4 w-4 text-purple-600" />
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold text-gray-900">{apiKeysLoading ? '...' : (apiKeysData?.length || 0)}</div>
                  <p className="text-xs text-gray-500">Active keys</p>
                </CardContent>
              </Card>

              <Card className="card-hover">
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                  <CardTitle className="text-sm font-medium text-gray-600">Total Cost</CardTitle>
                  <DollarSign className="h-4 w-4 text-green-600" />
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold text-gray-900">${usageLoading ? '...' : (usageData?.total_cost || 0).toFixed(2)}</div>
                  <p className="text-xs text-gray-500">This month</p>
                </CardContent>
              </Card>

              <Card className="card-hover">
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                  <CardTitle className="text-sm font-medium text-gray-600">Total Tokens</CardTitle>
                  <Database className="h-4 w-4 text-blue-600" />
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold text-gray-900">{usageLoading ? '...' : (usageData?.total_tokens || 0).toLocaleString()}</div>
                  <p className="text-xs text-gray-500">Processed tokens</p>
                </CardContent>
              </Card>
            </div>

            {/* Main Content Grid */}
            <div className="grid gap-8 lg:grid-cols-3">
              {/* Left Column - 2/3 width */}
              <div className="lg:col-span-2 space-y-8">
                {/* Recent Activity */}
                <Card>
                  <CardHeader>
                    <CardTitle className="text-lg font-semibold text-gray-900">Recent Activity</CardTitle>
                    <CardDescription className="text-gray-600">Your latest API requests and transactions</CardDescription>
                  </CardHeader>
                  <CardContent>
                    <div className="space-y-4">
                      <div className="flex items-center justify-between rounded-lg border border-gray-200 p-4 hover:bg-gray-50">
                        <div className="flex items-center space-x-3">
                          <div className="rounded-full bg-blue-100 p-2">
                            <Activity className="h-4 w-4 text-blue-600" />
                          </div>
                          <div>
                            <p className="font-medium text-gray-900">API Request</p>
                            <p className="text-sm text-gray-500">GPT-4o completion • 2 min ago</p>
                          </div>
                        </div>
                        <div className="text-right">
                          <span className="font-medium text-gray-900">$0.023</span>
                          <p className="text-xs text-gray-500">200 tokens</p>
                        </div>
                      </div>
                      <div className="flex items-center justify-between rounded-lg border border-gray-200 p-4 hover:bg-gray-50">
                        <div className="flex items-center space-x-3">
                          <div className="rounded-full bg-purple-100 p-2">
                            <Activity className="h-4 w-4 text-purple-600" />
                          </div>
                          <div>
                            <p className="font-medium text-gray-900">API Request</p>
                            <p className="text-sm text-gray-500">Claude 3 Sonnet • 15 min ago</p>
                          </div>
                        </div>
                        <div className="text-right">
                          <span className="font-medium text-gray-900">$0.015</span>
                          <p className="text-xs text-gray-500">150 tokens</p>
                        </div>
                      </div>
                      <div className="flex items-center justify-between rounded-lg border border-gray-200 p-4 hover:bg-gray-50">
                        <div className="flex items-center space-x-3">
                          <div className="rounded-full bg-green-100 p-2">
                            <CreditCard className="h-4 w-4 text-green-600" />
                          </div>
                          <div>
                            <p className="font-medium text-gray-900">Payment Received</p>
                            <p className="text-sm text-gray-500">Credit added • 1 hour ago</p>
                          </div>
                        </div>
                        <div className="text-right">
                          <span className="font-medium text-green-600">+$50.00</span>
                          <p className="text-xs text-gray-500">Balance updated</p>
                        </div>
                      </div>
                    </div>
                  </CardContent>
                </Card>

                {/* Top Used Models */}
                <Card>
                  <CardHeader>
                    <CardTitle className="text-lg font-semibold text-gray-900">Top Used Models</CardTitle>
                    <CardDescription className="text-gray-600">Your most frequently used AI models this month</CardDescription>
                  </CardHeader>
                  <CardContent>
                    {usageLoading ? (
                      <div className="flex justify-center py-8">
                        <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent"></div>
                      </div>
                    ) : usageData?.top_models?.length > 0 ? (
                      <div className="space-y-4">
                        {usageData.top_models.slice(0, 5).map((model: any, index: number) => (
                          <div key={index} className="flex items-center justify-between rounded-lg border border-gray-200 p-4 hover:bg-gray-50">
                            <div>
                              <p className="font-medium text-gray-900">{model.model_name}</p>
                              <p className="text-sm text-gray-500">{model.requests} requests • {model.tokens.toLocaleString()} tokens</p>
                            </div>
                            <div className="text-right">
                              <span className="font-medium text-gray-900">${model.cost.toFixed(4)}</span>
                              <p className="text-xs text-gray-500">Average: ${(model.cost / model.requests).toFixed(4)}/req</p>
                            </div>
                          </div>
                        ))}
                      </div>
                    ) : (
                      <div className="text-center py-8 text-gray-500">
                        <p>No usage data available yet.</p>
                        <p className="text-sm mt-2">Start using the API to see statistics here.</p>
                      </div>
                    )}
                  </CardContent>
                </Card>
              </div>

              {/* Right Column - 1/3 width */}
              <div className="space-y-8">
                {/* Quick Actions */}
                <Card>
                  <CardHeader>
                    <CardTitle className="text-lg font-semibold text-gray-900">Quick Actions</CardTitle>
                    <CardDescription className="text-gray-600">Common tasks to manage your account</CardDescription>
                  </CardHeader>
                  <CardContent>
                    <div className="space-y-3">
                      <Button className="w-full justify-start bg-gradient-primary hover:opacity-90 text-white" onClick={() => window.location.href = '/models'}>
                        <Zap className="mr-2 h-4 w-4" />
                        Browse Models
                      </Button>
                      <Button className="w-full justify-start" variant="outline" onClick={() => window.location.href = '/api-keys'}>
                        <Key className="mr-2 h-4 w-4" />
                        Manage API Keys
                      </Button>
                      <Button className="w-full justify-start" variant="outline" onClick={() => window.location.href = '/billing'}>
                        <CreditCard className="mr-2 h-4 w-4" />
                        Add Credit
                      </Button>
                      <Button className="w-full justify-start" variant="outline" onClick={() => window.location.href = '/usage'}>
                        <BarChart3 className="mr-2 h-4 w-4" />
                        View Usage Analytics
                      </Button>
                    </div>
                  </CardContent>
                </Card>

                {/* Account Summary */}
                <Card>
                  <CardHeader>
                    <CardTitle className="text-lg font-semibold text-gray-900">Account Summary</CardTitle>
                    <CardDescription className="text-gray-600">Your account overview</CardDescription>
                  </CardHeader>
                  <CardContent>
                    <div className="space-y-4">
                      <div className="flex items-center justify-between">
                        <span className="text-sm text-gray-600">Member Since</span>
                        <span className="font-medium text-gray-900">{new Date(user?.created_at || Date.now()).toLocaleDateString()}</span>
                      </div>
                      <div className="flex items-center justify-between">
                        <span className="text-sm text-gray-600">Account Status</span>
                        <span className="rounded-full bg-green-100 px-3 py-1 text-xs font-medium text-green-800">Active</span>
                      </div>
                      <div className="flex items-center justify-between">
                        <span className="text-sm text-gray-600">API Requests Today</span>
                        <span className="font-medium text-gray-900">42</span>
                      </div>
                      <div className="flex items-center justify-between">
                        <span className="text-sm text-gray-600">Monthly Spending</span>
                        <span className="font-medium text-gray-900">${usageData?.total_cost?.toFixed(2) || '0.00'}</span>
                      </div>
                    </div>
                    <Button className="mt-6 w-full" variant="outline" onClick={logout}>
                      Sign Out
                    </Button>
                  </CardContent>
                </Card>
              </div>
            </div>
          </div>
        </main>

        {/* Footer */}
        <footer className="mt-12 border-t bg-white py-8">
          <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
            <div className="flex flex-col items-center justify-between md:flex-row">
              <div className="mb-4 md:mb-0">
                <div className="flex items-center">
                  <Zap className="h-6 w-6 text-primary" />
                  <span className="ml-2 text-lg font-bold text-gray-900">MassRouter</span>
                </div>
                <p className="mt-2 text-sm text-gray-600">
                  AI Model Aggregation Platform • {new Date().getFullYear()}
                </p>
              </div>
              <div className="flex space-x-6">
                <a href="/privacy" className="text-sm text-gray-600 hover:text-gray-900">Privacy Policy</a>
                <a href="/terms" className="text-sm text-gray-600 hover:text-gray-900">Terms of Service</a>
                <a href="/support" className="text-sm text-gray-600 hover:text-gray-900">Support</a>
                <a href="/docs" className="text-sm text-gray-600 hover:text-gray-900">Documentation</a>
              </div>
            </div>
          </div>
        </footer>
      </div>
    </ProtectedRoute>
  );
}
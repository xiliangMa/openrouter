'use client';

import ProtectedRoute from '@/components/protected-route';
import { useAuth } from '@/lib/auth';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { BarChart3, CreditCard, Users, Zap, DollarSign, Settings, Bell, ChevronDown, Shield, Server, Key } from 'lucide-react';

export default function DashboardPage() {
  const { user, logout } = useAuth();

  return (
    <ProtectedRoute>
      <div className="min-h-screen bg-gray-50">
        {/* Navigation */}
        <nav className="sticky top-0 z-50 border-b bg-white shadow-sm">
          <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
            <div className="flex h-16 justify-between">
              <div className="flex">
                <div className="flex flex-shrink-0 items-center">
                  <Shield className="h-8 w-8 text-primary" />
                  <span className="ml-2 text-xl font-bold text-gray-900">MassRouter Admin</span>
                </div>
                <div className="hidden sm:ml-6 sm:flex sm:space-x-8">
                  <a href="/admin/dashboard" className="inline-flex items-center border-b-2 border-primary px-1 pt-1 text-sm font-medium text-gray-900">
                    Dashboard
                  </a>
                  <a href="/admin/users" className="inline-flex items-center border-b-2 border-transparent px-1 pt-1 text-sm font-medium text-gray-500 hover:border-gray-300 hover:text-gray-700">
                    Users
                  </a>
                  <a href="/admin/models" className="inline-flex items-center border-b-2 border-transparent px-1 pt-1 text-sm font-medium text-gray-500 hover:border-gray-300 hover:text-gray-700">
                    Models
                  </a>
                  <a href="/admin/analytics" className="inline-flex items-center border-b-2 border-transparent px-1 pt-1 text-sm font-medium text-gray-500 hover:border-gray-300 hover:text-gray-700">
                    Analytics
                  </a>
                  <a href="/admin/settings" className="inline-flex items-center border-b-2 border-transparent px-1 pt-1 text-sm font-medium text-gray-500 hover:border-gray-300 hover:text-gray-700">
                    Settings
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
              <h1 className="text-3xl font-bold text-gray-900">Admin Dashboard</h1>
              <p className="mt-2 text-gray-600">
                Welcome back, <span className="font-medium text-primary">{user?.username}</span>. Here&apos;s what&apos;s happening with your platform.
              </p>
            </div>

            {/* Stats Grid */}
            <div className="mb-8 grid gap-6 sm:grid-cols-2 lg:grid-cols-4">
              <Card className="card-hover">
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                  <CardTitle className="text-sm font-medium text-gray-600">Total Users</CardTitle>
                  <Users className="h-4 w-4 text-primary" />
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold text-gray-900">1,847</div>
                  <p className="text-xs text-gray-500">+12% from last month</p>
                </CardContent>
              </Card>

              <Card className="card-hover">
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                  <CardTitle className="text-sm font-medium text-gray-600">API Requests</CardTitle>
                  <BarChart3 className="h-4 w-4 text-purple-600" />
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold text-gray-900">245.2K</div>
                  <p className="text-xs text-gray-500">+23% from last week</p>
                </CardContent>
              </Card>

              <Card className="card-hover">
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                  <CardTitle className="text-sm font-medium text-gray-600">Total Revenue</CardTitle>
                  <DollarSign className="h-4 w-4 text-green-600" />
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold text-gray-900">$12,847</div>
                  <p className="text-xs text-gray-500">+18% from last month</p>
                </CardContent>
              </Card>

              <Card className="card-hover">
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                  <CardTitle className="text-sm font-medium text-gray-600">System Health</CardTitle>
                  <Server className="h-4 w-4 text-blue-600" />
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold text-gray-900">99.8%</div>
                  <p className="text-xs text-gray-500">Uptime this month</p>
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
                    <CardDescription className="text-gray-600">Latest platform events and user actions</CardDescription>
                  </CardHeader>
                  <CardContent>
                    <div className="space-y-4">
                      <div className="flex items-center justify-between rounded-lg border border-gray-200 p-4 hover:bg-gray-50">
                        <div className="flex items-center space-x-3">
                          <div className="rounded-full bg-blue-100 p-2">
                            <Users className="h-4 w-4 text-blue-600" />
                          </div>
                          <div>
                            <p className="font-medium text-gray-900">New User Registration</p>
                            <p className="text-sm text-gray-500">John Doe • 2 minutes ago</p>
                          </div>
                        </div>
                        <div className="text-right">
                          <span className="rounded-full bg-green-100 px-3 py-1 text-xs font-medium text-green-800">Approved</span>
                        </div>
                      </div>
                      <div className="flex items-center justify-between rounded-lg border border-gray-200 p-4 hover:bg-gray-50">
                        <div className="flex items-center space-x-3">
                          <div className="rounded-full bg-purple-100 p-2">
                            <Key className="h-4 w-4 text-purple-600" />
                          </div>
                          <div>
                            <p className="font-medium text-gray-900">API Key Generated</p>
                            <p className="text-sm text-gray-500">Project X • 15 minutes ago</p>
                          </div>
                        </div>
                        <div className="text-right">
                          <span className="rounded-full bg-blue-100 px-3 py-1 text-xs font-medium text-blue-800">Active</span>
                        </div>
                      </div>
                      <div className="flex items-center justify-between rounded-lg border border-gray-200 p-4 hover:bg-gray-50">
                        <div className="flex items-center space-x-3">
                          <div className="rounded-full bg-green-100 p-2">
                            <CreditCard className="h-4 w-4 text-green-600" />
                          </div>
                          <div>
                            <p className="font-medium text-gray-900">Payment Processed</p>
                            <p className="text-sm text-gray-500">$50.00 • 1 hour ago</p>
                          </div>
                        </div>
                        <div className="text-right">
                          <span className="rounded-full bg-green-100 px-3 py-1 text-xs font-medium text-green-800">Completed</span>
                        </div>
                      </div>
                      <div className="flex items-center justify-between rounded-lg border border-gray-200 p-4 hover:bg-gray-50">
                        <div className="flex items-center space-x-3">
                          <div className="rounded-full bg-yellow-100 p-2">
                            <Server className="h-4 w-4 text-yellow-600" />
                          </div>
                          <div>
                            <p className="font-medium text-gray-900">System Alert</p>
                            <p className="text-sm text-gray-500">High API latency detected • 3 hours ago</p>
                          </div>
                        </div>
                        <div className="text-right">
                          <span className="rounded-full bg-yellow-100 px-3 py-1 text-xs font-medium text-yellow-800">Warning</span>
                        </div>
                      </div>
                    </div>
                  </CardContent>
                </Card>

                {/* Top Models Usage */}
                <Card>
                  <CardHeader>
                    <CardTitle className="text-lg font-semibold text-gray-900">Top Models Usage</CardTitle>
                    <CardDescription className="text-gray-600">Most frequently used AI models across the platform</CardDescription>
                  </CardHeader>
                  <CardContent>
                    <div className="space-y-4">
                      <div className="flex items-center justify-between rounded-lg border border-gray-200 p-4 hover:bg-gray-50">
                        <div>
                          <p className="font-medium text-gray-900">GPT-4o</p>
                          <p className="text-sm text-gray-500">OpenAI • 45,200 requests</p>
                        </div>
                        <div className="text-right">
                          <span className="font-medium text-gray-900">$2,845.60</span>
                          <p className="text-xs text-gray-500">Average: $0.063/req</p>
                        </div>
                      </div>
                      <div className="flex items-center justify-between rounded-lg border border-gray-200 p-4 hover:bg-gray-50">
                        <div>
                          <p className="font-medium text-gray-900">Claude 3 Sonnet</p>
                          <p className="text-sm text-gray-500">Anthropic • 28,500 requests</p>
                        </div>
                        <div className="text-right">
                          <span className="font-medium text-gray-900">$1,927.50</span>
                          <p className="text-xs text-gray-500">Average: $0.068/req</p>
                        </div>
                      </div>
                      <div className="flex items-center justify-between rounded-lg border border-gray-200 p-4 hover:bg-gray-50">
                        <div>
                          <p className="font-medium text-gray-900">Gemini Pro</p>
                          <p className="text-sm text-gray-500">Google • 19,300 requests</p>
                        </div>
                        <div className="text-right">
                          <span className="font-medium text-gray-900">$1,158.00</span>
                          <p className="text-xs text-gray-500">Average: $0.060/req</p>
                        </div>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              </div>

              {/* Right Column - 1/3 width */}
              <div className="space-y-8">
                {/* Quick Actions */}
                <Card>
                  <CardHeader>
                    <CardTitle className="text-lg font-semibold text-gray-900">Quick Actions</CardTitle>
                    <CardDescription className="text-gray-600">Common admin tasks</CardDescription>
                  </CardHeader>
                  <CardContent>
                    <div className="space-y-3">
                      <Button className="w-full justify-start bg-gradient-primary hover:opacity-90 text-white" onClick={() => window.location.href = '/admin/users'}>
                        <Users className="mr-2 h-4 w-4" />
                        Manage Users
                      </Button>
                      <Button className="w-full justify-start" variant="outline" onClick={() => window.location.href = '/admin/models'}>
                        <Zap className="mr-2 h-4 w-4" />
                        Manage Models
                      </Button>
                      <Button className="w-full justify-start" variant="outline" onClick={() => window.location.href = '/admin/analytics'}>
                        <BarChart3 className="mr-2 h-4 w-4" />
                        View Analytics
                      </Button>
                      <Button className="w-full justify-start" variant="outline" onClick={() => window.location.href = '/admin/settings'}>
                        <Settings className="mr-2 h-4 w-4" />
                        System Settings
                      </Button>
                      <Button className="w-full justify-start" variant="outline" onClick={() => window.location.href = '/admin/api-keys'}>
                        <Key className="mr-2 h-4 w-4" />
                        API Keys
                      </Button>
                    </div>
                  </CardContent>
                </Card>

                {/* System Status */}
                <Card>
                  <CardHeader>
                    <CardTitle className="text-lg font-semibold text-gray-900">System Status</CardTitle>
                    <CardDescription className="text-gray-600">Platform health and metrics</CardDescription>
                  </CardHeader>
                  <CardContent>
                    <div className="space-y-4">
                      <div className="flex items-center justify-between">
                        <span className="text-sm text-gray-600">API Uptime</span>
                        <span className="font-medium text-gray-900">99.8%</span>
                      </div>
                      <div className="flex items-center justify-between">
                        <span className="text-sm text-gray-600">Active Sessions</span>
                        <span className="font-medium text-gray-900">342</span>
                      </div>
                      <div className="flex items-center justify-between">
                        <span className="text-sm text-gray-600">Avg Response Time</span>
                        <span className="font-medium text-gray-900">128ms</span>
                      </div>
                      <div className="flex items-center justify-between">
                        <span className="text-sm text-gray-600">Error Rate</span>
                        <span className="font-medium text-gray-900">0.12%</span>
                      </div>
                      <div className="flex items-center justify-between">
                        <span className="text-sm text-gray-600">Database Load</span>
                        <span className="rounded-full bg-green-100 px-3 py-1 text-xs font-medium text-green-800">Low</span>
                      </div>
                    </div>
                  </CardContent>
                </Card>

                {/* Account Summary */}
                <Card>
                  <CardHeader>
                    <CardTitle className="text-lg font-semibold text-gray-900">Account Summary</CardTitle>
                    <CardDescription className="text-gray-600">Your admin account overview</CardDescription>
                  </CardHeader>
                  <CardContent>
                    <div className="space-y-4">
                      <div className="flex items-center justify-between">
                        <span className="text-sm text-gray-600">Role</span>
                        <span className="rounded-full bg-purple-100 px-3 py-1 text-xs font-medium text-purple-800">Administrator</span>
                      </div>
                      <div className="flex items-center justify-between">
                        <span className="text-sm text-gray-600">Last Login</span>
                        <span className="font-medium text-gray-900">Today, 09:42 AM</span>
                      </div>
                      <div className="flex items-center justify-between">
                        <span className="text-sm text-gray-600">Account Created</span>
                        <span className="font-medium text-gray-900">{new Date(user?.created_at || Date.now()).toLocaleDateString()}</span>
                      </div>
                      <Button className="mt-4 w-full" variant="outline" onClick={logout}>
                        Sign Out
                      </Button>
                    </div>
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
                  <Shield className="h-6 w-6 text-primary" />
                  <span className="ml-2 text-lg font-bold text-gray-900">MassRouter Admin</span>
                </div>
                <p className="mt-2 text-sm text-gray-600">
                  AI Model Aggregation Platform • {new Date().getFullYear()}
                </p>
              </div>
              <div className="flex space-x-6">
                <a href="/admin/privacy" className="text-sm text-gray-600 hover:text-gray-900">Privacy Policy</a>
                <a href="/admin/terms" className="text-sm text-gray-600 hover:text-gray-900">Terms of Service</a>
                <a href="/admin/support" className="text-sm text-gray-600 hover:text-gray-900">Support</a>
                <a href="/admin/docs" className="text-sm text-gray-600 hover:text-gray-900">Documentation</a>
              </div>
            </div>
          </div>
        </footer>
      </div>
    </ProtectedRoute>
  );
}
'use client';

import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import ProtectedRoute from '@/components/protected-route'

export default function Home() {
  return (
    <ProtectedRoute>
      <div className="min-h-screen flex flex-col items-center justify-center p-8">
        <div className="w-full max-w-6xl mx-auto">
        <header className="mb-12">
          <h1 className="text-4xl font-bold tracking-tight">MassRouter SaaS Platform</h1>
          <p className="text-muted-foreground mt-2">Admin Dashboard</p>
        </header>

        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          <Card>
            <CardHeader>
              <CardTitle>Users</CardTitle>
              <CardDescription>Manage platform users</CardDescription>
            </CardHeader>
            <CardContent>
              <p className="text-sm text-muted-foreground mb-4">
                View, edit, and manage user accounts, roles, and permissions.
              </p>
              <Button>Manage Users</Button>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Models</CardTitle>
              <CardDescription>Configure AI models</CardDescription>
            </CardHeader>
            <CardContent>
              <p className="text-sm text-muted-foreground mb-4">
                Add, update, and manage AI models from various providers.
              </p>
              <Button>Manage Models</Button>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Billing</CardTitle>
              <CardDescription>Monitor revenue and payments</CardDescription>
            </CardHeader>
            <CardContent>
              <p className="text-sm text-muted-foreground mb-4">
                Track payments, revenue, and user billing information.
              </p>
              <Button>View Billing</Button>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Analytics</CardTitle>
              <CardDescription>Platform usage statistics</CardDescription>
            </CardHeader>
            <CardContent>
              <p className="text-sm text-muted-foreground mb-4">
                View usage patterns, model performance, and platform metrics.
              </p>
              <Button>View Analytics</Button>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>System Config</CardTitle>
              <CardDescription>Platform configuration</CardDescription>
            </CardHeader>
            <CardContent>
              <p className="text-sm text-muted-foreground mb-4">
                Configure system settings, feature flags, and global options.
              </p>
              <Button>Configure System</Button>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>API Keys</CardTitle>
              <CardDescription>Manage API access</CardDescription>
            </CardHeader>
            <CardContent>
              <p className="text-sm text-muted-foreground mb-4">
                Generate and revoke API keys for platform access.
              </p>
              <Button>Manage API Keys</Button>
            </CardContent>
          </Card>
        </div>

        <div className="mt-12">
          <Card>
            <CardHeader>
              <CardTitle>Quick Stats</CardTitle>
              <CardDescription>Platform overview</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                <div className="text-center p-4 border rounded-lg">
                  <div className="text-2xl font-bold">1,234</div>
                  <div className="text-sm text-muted-foreground">Total Users</div>
                </div>
                <div className="text-center p-4 border rounded-lg">
                  <div className="text-2xl font-bold">567</div>
                  <div className="text-sm text-muted-foreground">Active Today</div>
                </div>
                <div className="text-center p-4 border rounded-lg">
                  <div className="text-2xl font-bold">89</div>
                  <div className="text-sm text-muted-foreground">Models</div>
                </div>
                <div className="text-center p-4 border rounded-lg">
                  <div className="text-2xl font-bold">$12,345</div>
                  <div className="text-sm text-muted-foreground">Revenue</div>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
      </div>
    </ProtectedRoute>
  )
}
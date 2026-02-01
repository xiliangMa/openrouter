'use client';

import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import ProtectedRoute from '@/components/protected-route';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Plus, Edit, Trash2, Search, Database, Key, Link, Activity } from 'lucide-react';
import api from '@/lib/api';

interface ModelProvider {
  id: string;
  name: string;
  api_base_url: string;
  api_key: string; // Note: API key should be masked in UI
  config: Record<string, any>;
  status: string;
  created_at: string;
  updated_at: string;
  models_count?: number;
}

export default function ModelProvidersPage() {
  const queryClient = useQueryClient();
  const [search, setSearch] = useState('');
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingProvider, setEditingProvider] = useState<ModelProvider | null>(null);

  const { data: providers, isLoading } = useQuery({
    queryKey: ['admin-model-providers'],
    queryFn: async () => {
      const response = await api.get('/models/providers');
      return response.data.data.providers || response.data.data;
    },
  });

  const createProviderMutation = useMutation({
    mutationFn: async (provider: Partial<ModelProvider>) => {
      const response = await api.post('/admin/model-providers', provider);
      return response.data.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-model-providers'] });
      setDialogOpen(false);
      setEditingProvider(null);
    },
  });

  const updateProviderMutation = useMutation({
    mutationFn: async (provider: Partial<ModelProvider>) => {
      const response = await api.put(`/admin/model-providers/${provider.id}`, provider);
      return response.data.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-model-providers'] });
      setDialogOpen(false);
      setEditingProvider(null);
    },
  });

  const deleteProviderMutation = useMutation({
    mutationFn: async (id: string) => {
      await api.delete(`/admin/model-providers/${id}`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-model-providers'] });
    },
  });

  const filteredProviders = providers?.filter((provider: ModelProvider) =>
    provider.name.toLowerCase().includes(search.toLowerCase()) ||
    provider.api_base_url.toLowerCase().includes(search.toLowerCase())
  ) || [];

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!editingProvider) return;
    
    if (editingProvider.id) {
      updateProviderMutation.mutate(editingProvider);
    } else {
      createProviderMutation.mutate(editingProvider);
    }
  };

  return (
    <ProtectedRoute requireAdmin={true}>
      <div className="p-6 space-y-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Model Providers</h1>
            <p className="text-muted-foreground">
              Manage AI model providers and their configurations
            </p>
          </div>
          <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
            <DialogTrigger asChild>
              <Button onClick={() => setEditingProvider(null)}>
                <Plus className="mr-2 h-4 w-4" />
                Add Provider
              </Button>
            </DialogTrigger>
            <DialogContent className="sm:max-w-[600px]">
              <DialogHeader>
                <DialogTitle>{editingProvider ? 'Edit Provider' : 'Add New Provider'}</DialogTitle>
                <DialogDescription>
                  {editingProvider ? 'Update provider details.' : 'Add a new AI model provider to the platform.'}
                </DialogDescription>
              </DialogHeader>
              <form onSubmit={handleSubmit}>
                <div className="grid gap-4 py-4">
                  <div className="grid gap-2">
                    <Label htmlFor="name">Provider Name</Label>
                    <Input
                      id="name"
                      value={editingProvider?.name || ''}
                      onChange={(e) => setEditingProvider(prev => prev ? { ...prev, name: e.target.value } : null)}
                      placeholder="e.g., OpenAI, Anthropic, Google"
                      required
                    />
                  </div>
                  <div className="grid gap-2">
                    <Label htmlFor="api_base_url">API Base URL</Label>
                    <Input
                      id="api_base_url"
                      value={editingProvider?.api_base_url || ''}
                      onChange={(e) => setEditingProvider(prev => prev ? { ...prev, api_base_url: e.target.value } : null)}
                      placeholder="https://api.openai.com/v1"
                      required
                    />
                  </div>
                  <div className="grid gap-2">
                    <Label htmlFor="api_key">API Key</Label>
                    <Input
                      id="api_key"
                      type="password"
                      value={editingProvider?.api_key || ''}
                      onChange={(e) => setEditingProvider(prev => prev ? { ...prev, api_key: e.target.value } : null)}
                      placeholder="sk-..."
                      required={!editingProvider?.id}
                    />
                  </div>
                  <div className="grid gap-2">
                    <Label htmlFor="status">Status</Label>
                    <Select
                      value={editingProvider?.status || 'active'}
                      onValueChange={(value) => setEditingProvider(prev => prev ? { ...prev, status: value } : null)}
                    >
                      <SelectTrigger>
                        <SelectValue placeholder="Select status" />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="active">Active</SelectItem>
                        <SelectItem value="inactive">Inactive</SelectItem>
                        <SelectItem value="maintenance">Maintenance</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                </div>
                <DialogFooter>
                  <Button type="button" variant="outline" onClick={() => setDialogOpen(false)}>
                    Cancel
                  </Button>
                  <Button type="submit" disabled={createProviderMutation.isPending || updateProviderMutation.isPending}>
                    {createProviderMutation.isPending || updateProviderMutation.isPending ? 'Saving...' : editingProvider ? 'Update Provider' : 'Add Provider'}
                  </Button>
                </DialogFooter>
              </form>
            </DialogContent>
          </Dialog>
        </div>

        <Card>
          <CardHeader>
            <CardTitle>AI Model Providers</CardTitle>
            <CardDescription>
              All model providers configured on the platform
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="flex items-center gap-4 mb-6">
              <div className="relative flex-1">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground h-4 w-4" />
                <Input
                  placeholder="Search providers..."
                  value={search}
                  onChange={(e) => setSearch(e.target.value)}
                  className="pl-10"
                />
              </div>
            </div>

            {isLoading ? (
              <div className="text-center py-8">Loading providers...</div>
            ) : (
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Name</TableHead>
                    <TableHead>API URL</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead>Models</TableHead>
                    <TableHead>Created</TableHead>
                    <TableHead>Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {filteredProviders.map((provider: ModelProvider) => (
                    <TableRow key={provider.id}>
                      <TableCell className="font-medium">
                        <div className="flex items-center gap-2">
                          <Database className="h-4 w-4 text-muted-foreground" />
                          {provider.name}
                        </div>
                      </TableCell>
                      <TableCell>
                        <div className="flex items-center gap-2">
                          <Link className="h-4 w-4 text-muted-foreground" />
                          <span className="truncate max-w-[200px]">{provider.api_base_url}</span>
                        </div>
                      </TableCell>
                      <TableCell>
                        <Badge variant={
                          provider.status === 'active' ? 'default' :
                          provider.status === 'inactive' ? 'secondary' : 'destructive'
                        }>
                          {provider.status}
                        </Badge>
                      </TableCell>
                      <TableCell>
                        <Badge variant="outline">
                          {provider.models_count || 0} models
                        </Badge>
                      </TableCell>
                      <TableCell>
                        {new Date(provider.created_at).toLocaleDateString()}
                      </TableCell>
                      <TableCell>
                        <div className="flex items-center gap-2">
                          <Button
                            size="sm"
                            variant="outline"
                            onClick={() => {
                              setEditingProvider(provider);
                              setDialogOpen(true);
                            }}
                          >
                            <Edit className="h-4 w-4" />
                          </Button>
                          <Button
                            size="sm"
                            variant="destructive"
                            onClick={() => {
                              if (confirm('Are you sure you want to delete this provider? This will also delete all associated models.')) {
                                deleteProviderMutation.mutate(provider.id);
                              }
                            }}
                          >
                            <Trash2 className="h-4 w-4" />
                          </Button>
                        </div>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            )}
          </CardContent>
        </Card>
      </div>
    </ProtectedRoute>
  );
}
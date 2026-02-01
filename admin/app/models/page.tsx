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
import { Plus, Edit, Trash2, Search, Database, Zap, DollarSign, Eye } from 'lucide-react';
import api from '@/lib/api';

interface Model {
  id: string;
  name: string;
  description: string;
  provider_name: string;
  category: string;
  context_length: number | null;
  max_tokens: number | null;
  input_price: number;
  output_price: number;
  is_free: boolean;
  capabilities: Record<string, any>;
  pricing_tier: string;
  provider: {
    id: string;
    name: string;
  };
}

export default function ModelsPage() {
  const queryClient = useQueryClient();
  const [search, setSearch] = useState('');
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingModel, setEditingModel] = useState<Model | null>(null);

  const { data: models, isLoading } = useQuery({
    queryKey: ['admin-models'],
    queryFn: async () => {
      const response = await api.get('/models');
      return response.data.data.models;
    },
  });

  const createModelMutation = useMutation({
    mutationFn: async (model: Partial<Model>) => {
      const response = await api.post('/admin/models', model);
      return response.data.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-models'] });
      setDialogOpen(false);
      setEditingModel(null);
    },
  });

  const updateModelMutation = useMutation({
    mutationFn: async (model: Partial<Model>) => {
      const response = await api.put(`/admin/models/${model.id}`, model);
      return response.data.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-models'] });
      setDialogOpen(false);
      setEditingModel(null);
    },
  });

  const deleteModelMutation = useMutation({
    mutationFn: async (id: string) => {
      await api.delete(`/admin/models/${id}`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-models'] });
    },
  });

  const filteredModels = models?.filter((model: Model) =>
    model.name.toLowerCase().includes(search.toLowerCase()) ||
    model.description?.toLowerCase().includes(search.toLowerCase()) ||
    model.provider_name.toLowerCase().includes(search.toLowerCase())
  ) || [];

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!editingModel) return;
    
    if (editingModel.id) {
      updateModelMutation.mutate(editingModel);
    } else {
      createModelMutation.mutate(editingModel);
    }
  };

  return (
    <ProtectedRoute requireAdmin={true}>
      <div className="p-6 space-y-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Model Management</h1>
            <p className="text-muted-foreground">
              Manage AI models available on the platform
            </p>
          </div>
          <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
            <DialogTrigger asChild>
              <Button onClick={() => setEditingModel(null)}>
                <Plus className="mr-2 h-4 w-4" />
                Add Model
              </Button>
            </DialogTrigger>
            <DialogContent className="sm:max-w-[600px]">
              <DialogHeader>
                <DialogTitle>{editingModel ? 'Edit Model' : 'Add New Model'}</DialogTitle>
                <DialogDescription>
                  {editingModel ? 'Update model details.' : 'Add a new AI model to the platform.'}
                </DialogDescription>
              </DialogHeader>
              <form onSubmit={handleSubmit}>
                <div className="grid gap-4 py-4">
                  <div className="grid gap-2">
                    <Label htmlFor="name">Model Name</Label>
                    <Input
                      id="name"
                      value={editingModel?.name || ''}
                      onChange={(e) => setEditingModel(prev => prev ? { ...prev, name: e.target.value } : null)}
                      placeholder="e.g., gpt-4-turbo"
                    />
                  </div>
                  <div className="grid gap-2">
                    <Label htmlFor="description">Description</Label>
                    <Input
                      id="description"
                      value={editingModel?.description || ''}
                      onChange={(e) => setEditingModel(prev => prev ? { ...prev, description: e.target.value } : null)}
                      placeholder="Model description"
                    />
                  </div>
                  <div className="grid gap-2">
                    <Label htmlFor="provider_name">Provider Name</Label>
                    <Input
                      id="provider_name"
                      value={editingModel?.provider_name || ''}
                      onChange={(e) => setEditingModel(prev => prev ? { ...prev, provider_name: e.target.value } : null)}
                      placeholder="e.g., OpenAI"
                    />
                  </div>
                  <div className="grid gap-2">
                    <Label htmlFor="input_price">Input Price (per 1M tokens)</Label>
                    <Input
                      id="input_price"
                      type="number"
                      step="0.000001"
                      value={editingModel?.input_price || ''}
                      onChange={(e) => setEditingModel(prev => prev ? { ...prev, input_price: parseFloat(e.target.value) } : null)}
                    />
                  </div>
                  <div className="grid gap-2">
                    <Label htmlFor="output_price">Output Price (per 1M tokens)</Label>
                    <Input
                      id="output_price"
                      type="number"
                      step="0.000001"
                      value={editingModel?.output_price || ''}
                      onChange={(e) => setEditingModel(prev => prev ? { ...prev, output_price: parseFloat(e.target.value) } : null)}
                    />
                  </div>
                </div>
                <DialogFooter>
                  <Button type="button" variant="outline" onClick={() => setDialogOpen(false)}>
                    Cancel
                  </Button>
                  <Button type="submit" disabled={updateModelMutation.isPending}>
                    {updateModelMutation.isPending ? 'Saving...' : editingModel ? 'Update Model' : 'Add Model'}
                  </Button>
                </DialogFooter>
              </form>
            </DialogContent>
          </Dialog>
        </div>

        <Card>
          <CardHeader>
            <CardTitle>AI Models</CardTitle>
            <CardDescription>
              All models available on the platform
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="flex items-center gap-4 mb-6">
              <div className="relative flex-1">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground h-4 w-4" />
                <Input
                  placeholder="Search models..."
                  value={search}
                  onChange={(e) => setSearch(e.target.value)}
                  className="pl-10"
                />
              </div>
            </div>

            {isLoading ? (
              <div className="text-center py-8">Loading models...</div>
            ) : (
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Name</TableHead>
                    <TableHead>Provider</TableHead>
                    <TableHead>Category</TableHead>
                    <TableHead>Input/Output Price</TableHead>
                    <TableHead>Free</TableHead>
                    <TableHead>Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {filteredModels.map((model: Model) => (
                    <TableRow key={model.id}>
                      <TableCell className="font-medium">{model.name}</TableCell>
                      <TableCell>
                        <Badge variant="outline">{model.provider_name}</Badge>
                      </TableCell>
                      <TableCell>{model.category || '-'}</TableCell>
                      <TableCell>
                        <div className="flex flex-col">
                          <span>${model.input_price?.toFixed(6)}/M input</span>
                          <span>${model.output_price?.toFixed(6)}/M output</span>
                        </div>
                      </TableCell>
                      <TableCell>
                        <Badge variant={model.is_free ? "default" : "secondary"}>
                          {model.is_free ? 'Yes' : 'No'}
                        </Badge>
                      </TableCell>
                      <TableCell>
                        <div className="flex items-center gap-2">
                          <Button
                            size="sm"
                            variant="outline"
                            onClick={() => {
                              setEditingModel(model);
                              setDialogOpen(true);
                            }}
                          >
                            <Edit className="h-4 w-4" />
                          </Button>
                          <Button
                            size="sm"
                            variant="destructive"
                            onClick={() => {
                              if (confirm('Are you sure you want to delete this model?')) {
                                deleteModelMutation.mutate(model.id);
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
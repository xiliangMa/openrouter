'use client';

import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import ProtectedRoute from '@/components/protected-route';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Plus, Copy, Trash2, Key, Clock, Check, X, Eye, EyeOff } from 'lucide-react';
import api from '@/lib/api';
import { toast } from 'sonner';

interface APIKey {
  id: string;
  name: string;
  api_key: string;
  prefix: string;
  permissions: string[];
  rate_limit: number;
  expires_at: string | null;
  last_used_at: string | null;
  is_active: boolean;
  created_at: string;
}

export default function APIKeysPage() {
  const queryClient = useQueryClient();
  const [dialogOpen, setDialogOpen] = useState(false);
  const [newKeyName, setNewKeyName] = useState('');
  const [newKeyPermissions, setNewKeyPermissions] = useState<string[]>(['read', 'write']);
  const [newKeyRateLimit, setNewKeyRateLimit] = useState('1000');
  const [newKeyExpiresIn, setNewKeyExpiresIn] = useState('30');
  const [showKey, setShowKey] = useState<string | null>(null);
  const [copiedKeyId, setCopiedKeyId] = useState<string | null>(null);

  const { data: apiKeys, isLoading } = useQuery({
    queryKey: ['user-api-keys'],
    queryFn: async () => {
      const response = await api.get('/user/api-keys');
      return response.data.data.api_keys || [];
    },
  });

  const createKeyMutation = useMutation({
    mutationFn: async (keyData: any) => {
      const response = await api.post('/user/api-keys', keyData);
      return response.data.data;
    },
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ['user-api-keys'] });
      setDialogOpen(false);
      setNewKeyName('');
      setNewKeyPermissions(['read', 'write']);
      setNewKeyRateLimit('1000');
      setNewKeyExpiresIn('30');
      
      // Show the new API key
      setShowKey(data.api_key);
      toast.success('API key created successfully!');
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.error?.message || 'Failed to create API key');
    },
  });

  const deleteKeyMutation = useMutation({
    mutationFn: async (keyId: string) => {
      await api.delete(`/user/api-keys/${keyId}`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['user-api-keys'] });
      toast.success('API key deleted successfully!');
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.error?.message || 'Failed to delete API key');
    },
  });

  const toggleKeyMutation = useMutation({
    mutationFn: async ({ keyId, active }: { keyId: string; active: boolean }) => {
      // Note: This endpoint doesn't exist yet, would need backend implementation
      // For now, we'll just delete and recreate or use a different approach
      // We'll implement toggle by showing a message that this feature is coming soon
      throw new Error('Toggle feature coming soon');
    },
  });

  const copyToClipboard = async (text: string, keyId: string) => {
    try {
      await navigator.clipboard.writeText(text);
      setCopiedKeyId(keyId);
      toast.success('API key copied to clipboard!');
      setTimeout(() => setCopiedKeyId(null), 2000);
    } catch (err) {
      toast.error('Failed to copy to clipboard');
    }
  };

  const handleCreateKey = (e: React.FormEvent) => {
    e.preventDefault();
    if (!newKeyName.trim()) {
      toast.error('Please enter a name for the API key');
      return;
    }

    createKeyMutation.mutate({
      name: newKeyName,
      permissions: newKeyPermissions,
      rate_limit: parseInt(newKeyRateLimit),
      expires_in: parseInt(newKeyExpiresIn) * 24 * 3600, // Convert days to seconds
    });
  };

  const handleDeleteKey = (keyId: string) => {
    if (confirm('Are you sure you want to delete this API key? This action cannot be undone.')) {
      deleteKeyMutation.mutate(keyId);
    }
  };

  const formatDate = (dateString: string | null) => {
    if (!dateString) return 'Never';
    return new Date(dateString).toLocaleString();
  };

  const isExpired = (expiresAt: string | null) => {
    if (!expiresAt) return false;
    return new Date(expiresAt) < new Date();
  };

  return (
    <ProtectedRoute>
      <div className="min-h-screen bg-gray-50">
        <header className="sticky top-0 z-50 border-b bg-white">
          <div className="flex h-16 items-center justify-between px-6">
            <div className="flex items-center gap-3">
              <Key className="h-6 w-6 text-primary" />
              <h1 className="text-xl font-semibold">MassRouter Portal</h1>
            </div>
            <Button variant="outline" onClick={() => window.location.href = '/dashboard'}>
              Back to Dashboard
            </Button>
          </div>
        </header>

        <main className="p-6">
          <div className="mb-8">
            <h2 className="text-3xl font-bold tracking-tight">API Keys</h2>
            <p className="text-muted-foreground">
              Manage your API keys for accessing the MassRouter API.
            </p>
          </div>

          {/* Stats */}
          <div className="grid gap-6 mb-8 md:grid-cols-4">
            <Card>
              <CardContent className="pt-6">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm text-muted-foreground">Total Keys</p>
                    <p className="text-2xl font-bold">{apiKeys?.length || 0}</p>
                  </div>
                  <Key className="h-8 w-8 text-muted-foreground" />
                </div>
              </CardContent>
            </Card>
            <Card>
              <CardContent className="pt-6">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm text-muted-foreground">Active Keys</p>
                    <p className="text-2xl font-bold">
                      {apiKeys?.filter((k: APIKey) => k.is_active).length || 0}
                    </p>
                  </div>
                  <Check className="h-8 w-8 text-green-500" />
                </div>
              </CardContent>
            </Card>
            <Card>
              <CardContent className="pt-6">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm text-muted-foreground">Expired Keys</p>
                    <p className="text-2xl font-bold">
                      {apiKeys?.filter((k: APIKey) => isExpired(k.expires_at)).length || 0}
                    </p>
                  </div>
                  <X className="h-8 w-8 text-red-500" />
                </div>
              </CardContent>
            </Card>
            <Card>
              <CardContent className="pt-6">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm text-muted-foreground">Last Created</p>
                    <p className="text-lg font-medium">
                      {apiKeys?.length 
                        ? new Date(apiKeys[0].created_at).toLocaleDateString() 
                        : 'Never'}
                    </p>
                  </div>
                  <Clock className="h-8 w-8 text-muted-foreground" />
                </div>
              </CardContent>
            </Card>
          </div>

          {/* Create Key Dialog */}
          <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
            <div className="flex justify-between items-center mb-6">
              <h3 className="text-lg font-medium">Your API Keys</h3>
              <DialogTrigger asChild>
                <Button>
                  <Plus className="mr-2 h-4 w-4" />
                  Create New Key
                </Button>
              </DialogTrigger>
            </div>
            
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Create New API Key</DialogTitle>
                <DialogDescription>
                  Create a new API key to access the MassRouter API.
                </DialogDescription>
              </DialogHeader>
              
              <form onSubmit={handleCreateKey}>
                <div className="grid gap-4 py-4">
                  <div className="grid gap-2">
                    <Label htmlFor="name">Key Name</Label>
                    <Input
                      id="name"
                      placeholder="e.g., Production Server, Development"
                      value={newKeyName}
                      onChange={(e) => setNewKeyName(e.target.value)}
                    />
                  </div>
                  
                  <div className="grid gap-2">
                    <Label htmlFor="permissions">Permissions</Label>
                    <Select
                      value={newKeyPermissions.join(',')}
                       onValueChange={(value: string) => setNewKeyPermissions(value.split(','))}
                    >
                      <SelectTrigger>
                        <SelectValue placeholder="Select permissions" />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="read">Read Only</SelectItem>
                        <SelectItem value="read,write">Read & Write</SelectItem>
                        <SelectItem value="read,write,admin">Full Access</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                  
                  <div className="grid gap-2">
                    <Label htmlFor="rate_limit">Rate Limit (requests/minute)</Label>
                    <Input
                      id="rate_limit"
                      type="number"
                      min="1"
                      max="10000"
                      value={newKeyRateLimit}
                      onChange={(e) => setNewKeyRateLimit(e.target.value)}
                    />
                  </div>
                  
                  <div className="grid gap-2">
                    <Label htmlFor="expires_in">Expires In (days)</Label>
                    <Select value={newKeyExpiresIn} onValueChange={setNewKeyExpiresIn}>
                      <SelectTrigger>
                        <SelectValue placeholder="Select expiration" />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="7">7 days</SelectItem>
                        <SelectItem value="30">30 days</SelectItem>
                        <SelectItem value="90">90 days</SelectItem>
                        <SelectItem value="365">1 year</SelectItem>
                        <SelectItem value="never">Never</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                </div>
                
                <DialogFooter>
                  <Button
                    type="button"
                    variant="outline"
                    onClick={() => setDialogOpen(false)}
                  >
                    Cancel
                  </Button>
                  <Button type="submit" disabled={createKeyMutation.isPending}>
                    {createKeyMutation.isPending ? 'Creating...' : 'Create Key'}
                  </Button>
                </DialogFooter>
              </form>
            </DialogContent>
          </Dialog>

          {/* New Key Display Dialog */}
          {showKey && (
             <Dialog open={!!showKey} onOpenChange={(open: boolean) => !open && setShowKey(null)}>
              <DialogContent>
                <DialogHeader>
                  <DialogTitle>API Key Created Successfully!</DialogTitle>
                  <DialogDescription>
                    Copy your new API key now. You won't be able to see it again.
                  </DialogDescription>
                </DialogHeader>
                
                <div className="py-4">
                  <Label htmlFor="new-key">Your API Key</Label>
                  <div className="flex items-center gap-2 mt-2">
                    <Input
                      id="new-key"
                      value={showKey}
                      readOnly
                      type={copiedKeyId === 'new' ? 'text' : 'password'}
                    />
                    <Button
                      size="icon"
                      variant="outline"
                      onClick={() => copyToClipboard(showKey, 'new')}
                    >
                      {copiedKeyId === 'new' ? (
                        <Check className="h-4 w-4 text-green-600" />
                      ) : (
                        <Copy className="h-4 w-4" />
                      )}
                    </Button>
                    <Button
                      size="icon"
                      variant="outline"
                      onClick={() => setShowKey(null)}
                    >
                      <EyeOff className="h-4 w-4" />
                    </Button>
                  </div>
                  <p className="text-sm text-muted-foreground mt-2">
                    ⚠️ Store this key securely. It will not be shown again.
                  </p>
                </div>
                
                <DialogFooter>
                  <Button onClick={() => setShowKey(null)}>Done</Button>
                </DialogFooter>
              </DialogContent>
            </Dialog>
          )}

          {/* API Keys Table */}
          <Card>
            <CardContent className="pt-6">
              {isLoading ? (
                <div className="flex items-center justify-center py-12">
                  <div className="text-center">
                    <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent mx-auto"></div>
                    <p className="mt-4 text-muted-foreground">Loading API keys...</p>
                  </div>
                </div>
              ) : apiKeys?.length === 0 ? (
                <div className="text-center py-12">
                  <Key className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
                  <h3 className="text-lg font-medium">No API keys yet</h3>
                  <p className="text-muted-foreground mt-2">
                    Create your first API key to start using the MassRouter API.
                  </p>
                  <Button className="mt-4" onClick={() => setDialogOpen(true)}>
                    <Plus className="mr-2 h-4 w-4" />
                    Create First Key
                  </Button>
                </div>
              ) : (
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Name</TableHead>
                      <TableHead>Key</TableHead>
                      <TableHead>Permissions</TableHead>
                      <TableHead>Rate Limit</TableHead>
                      <TableHead>Status</TableHead>
                      <TableHead>Expires</TableHead>
                      <TableHead>Last Used</TableHead>
                      <TableHead>Actions</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {apiKeys?.map((key: APIKey) => (
                      <TableRow key={key.id}>
                        <TableCell className="font-medium">{key.name}</TableCell>
                        <TableCell>
                          <div className="flex items-center gap-2">
                            <code className="text-xs bg-muted px-2 py-1 rounded">
                              {key.prefix}•••••••
                            </code>
                            <Button
                              size="icon"
                              variant="ghost"
                              className="h-6 w-6"
                              onClick={() => copyToClipboard(key.api_key, key.id)}
                            >
                              {copiedKeyId === key.id ? (
                                <Check className="h-3 w-3 text-green-600" />
                              ) : (
                                <Copy className="h-3 w-3" />
                              )}
                            </Button>
                          </div>
                        </TableCell>
                        <TableCell>
                          <div className="flex flex-wrap gap-1">
                            {key.permissions?.map((perm) => (
                              <Badge key={perm} variant="outline" className="text-xs">
                                {perm}
                              </Badge>
                            ))}
                          </div>
                        </TableCell>
                        <TableCell>{key.rate_limit}/min</TableCell>
                        <TableCell>
                          <Badge
                            variant={key.is_active ? "default" : "secondary"}
                            className={isExpired(key.expires_at) ? "bg-red-100 text-red-800" : ""}
                          >
                            {key.is_active 
                              ? isExpired(key.expires_at) ? 'Expired' : 'Active' 
                              : 'Inactive'}
                          </Badge>
                        </TableCell>
                        <TableCell>
                          {key.expires_at ? formatDate(key.expires_at) : 'Never'}
                        </TableCell>
                        <TableCell>{formatDate(key.last_used_at)}</TableCell>
                        <TableCell>
                          <div className="flex items-center gap-2">
                            <Button
                              size="sm"
                              variant="outline"
                              onClick={() => handleDeleteKey(key.id)}
                              disabled={deleteKeyMutation.isPending}
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

          {/* Usage Instructions */}
          <Card className="mt-8">
            <CardHeader>
              <CardTitle>Using Your API Keys</CardTitle>
              <CardDescription>How to authenticate with the MassRouter API</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                <div>
                  <h4 className="font-medium mb-2">Authentication Header</h4>
                  <code className="block text-sm bg-muted p-3 rounded mb-2">
                    Authorization: Bearer YOUR_API_KEY
                  </code>
                  <p className="text-sm text-muted-foreground">
                    Include this header in all API requests.
                  </p>
                </div>
                <div>
                  <h4 className="font-medium mb-2">Example cURL Request</h4>
                  <code className="block text-sm bg-muted p-3 rounded mb-2">
                    curl -X POST https://api.massrouter.ai/v1/chat/completions \<br />
                    &nbsp;&nbsp;-H "Authorization: Bearer YOUR_API_KEY" \<br />
                    &nbsp;&nbsp;-H "Content-Type: application/json" \<br />
                    &nbsp;&nbsp;-d '&#123;"model": "gpt-3.5-turbo", "messages": [&#123;"role": "user", "content": "Hello!"&#125;]&#125;'
                  </code>
                </div>
                <div>
                  <h4 className="font-medium mb-2">Security Best Practices</h4>
                  <ul className="text-sm text-muted-foreground list-disc list-inside space-y-1">
                    <li>Never commit API keys to version control</li>
                    <li>Use environment variables in production</li>
                    <li>Rotate keys periodically</li>
                    <li>Set appropriate rate limits for each use case</li>
                    <li>Delete unused keys</li>
                  </ul>
                </div>
              </div>
            </CardContent>
          </Card>
        </main>
      </div>
    </ProtectedRoute>
  );
}
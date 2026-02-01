'use client';

import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import Link from 'next/link';
import ProtectedRoute from '@/components/protected-route';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Search, Filter, Zap, Eye, DollarSign, Database, ChevronRight } from 'lucide-react';
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
  const [search, setSearch] = useState('');
  const [category, setCategory] = useState<string>('all');
  const [provider, setProvider] = useState<string>('all');
  const [priceFilter, setPriceFilter] = useState<string>('all');

  const { data: modelsData, isLoading } = useQuery({
    queryKey: ['models'],
    queryFn: async () => {
      const response = await api.get('/models');
      return response.data.data;
    },
  });

  const models = modelsData?.models || [];
  const providers: string[] = Array.from(new Set(models.map((m: Model) => m.provider_name))).sort() as string[];
  const categories: string[] = Array.from(new Set(models.map((m: Model) => m.category).filter(Boolean))).sort() as string[];

  const filteredModels = models.filter((model: Model) => {
    const matchesSearch = model.name.toLowerCase().includes(search.toLowerCase()) ||
                         model.description?.toLowerCase().includes(search.toLowerCase()) ||
                         model.provider_name.toLowerCase().includes(search.toLowerCase());
    
    const matchesCategory = category === 'all' || model.category === category;
    const matchesProvider = provider === 'all' || model.provider_name === provider;
    
    let matchesPrice = true;
    if (priceFilter === 'free') {
      matchesPrice = model.is_free === true;
    } else if (priceFilter === 'paid') {
      matchesPrice = model.is_free === false;
    } else if (priceFilter === 'economy') {
      matchesPrice = model.input_price < 0.001;
    } else if (priceFilter === 'premium') {
      matchesPrice = model.input_price >= 0.005;
    }
    
    return matchesSearch && matchesCategory && matchesProvider && matchesPrice;
  });

  const formatPrice = (price: number) => {
    if (price === 0) return 'Free';
    return `$${price.toFixed(6)}/1K`;
  };

  const getPriceColor = (price: number) => {
    if (price === 0) return 'text-green-600';
    if (price < 0.001) return 'text-blue-600';
    if (price < 0.005) return 'text-yellow-600';
    return 'text-purple-600';
  };

  return (
    <ProtectedRoute>
      <div className="min-h-screen bg-gray-50">
        <header className="sticky top-0 z-50 border-b bg-white">
          <div className="flex h-16 items-center justify-between px-6">
            <div className="flex items-center gap-3">
              <Zap className="h-6 w-6 text-primary" />
              <h1 className="text-xl font-semibold">MassRouter Portal</h1>
            </div>
            <Button variant="outline" onClick={() => window.location.href = '/dashboard'}>
              Back to Dashboard
            </Button>
          </div>
        </header>

        <main className="p-6">
          <div className="mb-8">
            <h2 className="text-3xl font-bold tracking-tight">AI Models</h2>
            <p className="text-muted-foreground">
              Browse and search through {models.length} AI models from {providers.length} providers.
            </p>
          </div>

          {/* Filters */}
          <Card className="mb-8">
            <CardContent className="pt-6">
              <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-5">
                <div className="relative">
                  <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                  <Input
                    placeholder="Search models..."
                    className="pl-9"
                    value={search}
                    onChange={(e) => setSearch(e.target.value)}
                  />
                </div>
                
                <Select value={category} onValueChange={setCategory}>
                  <SelectTrigger>
                    <SelectValue placeholder="Category" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">All Categories</SelectItem>
                     {categories.map((cat: string) => (
                      <SelectItem key={cat} value={cat}>{cat}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>

                <Select value={provider} onValueChange={setProvider}>
                  <SelectTrigger>
                    <SelectValue placeholder="Provider" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">All Providers</SelectItem>
                     {providers.map((prov: string) => (
                      <SelectItem key={prov} value={prov}>{prov}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>

                <Select value={priceFilter} onValueChange={setPriceFilter}>
                  <SelectTrigger>
                    <SelectValue placeholder="Price" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">All Prices</SelectItem>
                    <SelectItem value="free">Free Only</SelectItem>
                    <SelectItem value="paid">Paid Only</SelectItem>
                    <SelectItem value="economy">Economy (&lt; $0.001)</SelectItem>
                    <SelectItem value="premium">Premium (&gt;= $0.005)</SelectItem>
                  </SelectContent>
                </Select>

                <Button onClick={() => {
                  setSearch('');
                  setCategory('all');
                  setProvider('all');
                  setPriceFilter('all');
                }}>
                  <Filter className="mr-2 h-4 w-4" />
                  Clear Filters
                </Button>
              </div>
            </CardContent>
          </Card>

          {/* Models Grid */}
          {isLoading ? (
            <div className="flex items-center justify-center py-12">
              <div className="text-center">
                <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent mx-auto"></div>
                <p className="mt-4 text-muted-foreground">Loading models...</p>
              </div>
            </div>
          ) : (
            <>
              <div className="mb-4 flex items-center justify-between">
                <p className="text-sm text-muted-foreground">
                  Showing {filteredModels.length} of {models.length} models
                </p>
              </div>

              <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
                {filteredModels.map((model: Model) => (
                  <Card key={model.id} className="hover:shadow-lg transition-shadow">
                    <CardHeader>
                      <div className="flex items-start justify-between">
                        <div>
                          <CardTitle className="text-lg">{model.name}</CardTitle>
                          <CardDescription>{model.provider_name}</CardDescription>
                        </div>
                        <Badge variant={model.is_free ? "default" : "secondary"}>
                          {model.is_free ? 'Free' : 'Paid'}
                        </Badge>
                      </div>
                    </CardHeader>
                    <CardContent>
                      <div className="space-y-4">
                        <p className="text-sm text-muted-foreground line-clamp-2">
                          {model.description || 'No description available.'}
                        </p>
                        
                        <div className="grid grid-cols-2 gap-4">
                          <div>
                            <p className="text-xs text-muted-foreground">Input Price</p>
                            <p className={`font-medium ${getPriceColor(model.input_price)}`}>
                              {formatPrice(model.input_price)}
                            </p>
                          </div>
                          <div>
                            <p className="text-xs text-muted-foreground">Output Price</p>
                            <p className={`font-medium ${getPriceColor(model.output_price)}`}>
                              {formatPrice(model.output_price)}
                            </p>
                          </div>
                          <div>
                            <p className="text-xs text-muted-foreground">Context Length</p>
                            <p className="font-medium">
                              {model.context_length ? `${model.context_length.toLocaleString()} tokens` : 'N/A'}
                            </p>
                          </div>
                          <div>
                            <p className="text-xs text-muted-foreground">Max Tokens</p>
                            <p className="font-medium">
                              {model.max_tokens ? `${model.max_tokens.toLocaleString()} tokens` : 'N/A'}
                            </p>
                          </div>
                        </div>

                        {model.capabilities && Object.keys(model.capabilities).length > 0 && (
                          <div>
                            <p className="text-xs text-muted-foreground mb-2">Capabilities</p>
                            <div className="flex flex-wrap gap-1">
                              {Object.entries(model.capabilities).map(([key, value]) => (
                                value === true && (
                                  <Badge key={key} variant="outline" className="text-xs">
                                    {key}
                                  </Badge>
                                )
                              ))}
                            </div>
                          </div>
                        )}

                        <div className="flex items-center justify-between pt-4">
                          <div className="text-sm">
                            <Badge variant="outline">{model.category || 'Uncategorized'}</Badge>
                            {model.pricing_tier && (
                              <Badge variant="outline" className="ml-2">
                                {model.pricing_tier}
                              </Badge>
                            )}
                          </div>
                          <Link href={`/models/${model.id}`}>
                            <Button size="sm">
                              <Eye className="mr-2 h-3 w-3" />
                              View Details
                            </Button>
                          </Link>
                        </div>
                      </div>
                    </CardContent>
                  </Card>
                ))}
              </div>

              {filteredModels.length === 0 && (
                <div className="text-center py-12">
                  <Search className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
                  <h3 className="text-lg font-medium">No models found</h3>
                  <p className="text-muted-foreground mt-2">
                    Try adjusting your search or filter criteria.
                  </p>
                </div>
              )}
            </>
          )}

          {/* Provider Summary */}
          <Card className="mt-8">
            <CardHeader>
              <CardTitle>Model Providers</CardTitle>
              <CardDescription>Available AI model providers</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
                 {providers.map((providerName: string) => {
                  const providerModels = models.filter((m: Model) => m.provider_name === providerName);
                  const freeModels = providerModels.filter((m: Model) => m.is_free).length;
                  
                  return (
                    <Card key={providerName} className="border">
                      <CardContent className="pt-6">
                        <div className="flex items-center justify-between">
                          <div>
                            <h4 className="font-medium">{providerName}</h4>
                            <p className="text-sm text-muted-foreground">
                              {providerModels.length} models • {freeModels} free
                            </p>
                          </div>
                          <ChevronRight className="h-4 w-4 text-muted-foreground" />
                        </div>
                      </CardContent>
                    </Card>
                  );
                })}
              </div>
            </CardContent>
          </Card>
        </main>
      </div>
    </ProtectedRoute>
  );
}
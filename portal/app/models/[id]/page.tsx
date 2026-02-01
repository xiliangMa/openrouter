/* eslint-disable react/no-unescaped-entities */
'use client'

import { useState, useEffect } from 'react'
import { useParams, useRouter } from 'next/navigation'
import { useModelDetails } from '@/lib/hooks/useModels'
import { useCompare } from '@/lib/hooks/useCompare'
import ProtectedRoute from '@/components/protected-route'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog'
import { 
  ArrowLeft, 
  Zap, 
  DollarSign, 
  Database, 
  Cpu, 
  Shield, 
  Globe, 
  Code,
  BarChart,
  Clock,
  CheckCircle,
  XCircle,
  ChevronRight,
  Send,
  User,
  Bot,
  Loader2,
  Copy,
  Check
} from 'lucide-react'
import Link from 'next/link'
import { toast } from 'sonner'
import api from '@/lib/api'

export default function ModelDetailPage() {
  const params = useParams()
  const router = useRouter()
  const modelId = params.id as string
  
  const { data: modelData, isLoading, error } = useModelDetails(modelId)
  const model = modelData?.data
  const { addModel, removeModel, isInCompare, count } = useCompare()
  const isComparing = isInCompare(modelId)
  const [selectedTab, setSelectedTab] = useState<'curl' | 'python' | 'javascript'>('curl')
  const [tryModalOpen, setTryModalOpen] = useState(false)
  const [testMessage, setTestMessage] = useState('')
  const [testResponse, setTestResponse] = useState('')
  const [isTesting, setIsTesting] = useState(false)
  const [apiKeys, setApiKeys] = useState<any[]>([])
  const [selectedApiKey, setSelectedApiKey] = useState('')
  const [copied, setCopied] = useState(false)
  const [relatedModels, setRelatedModels] = useState<any[]>([])
  const [loadingRelated, setLoadingRelated] = useState(false)

  useEffect(() => {
    if (tryModalOpen) {
      fetchApiKeys()
    }
  }, [tryModalOpen])

  useEffect(() => {
    if (!model) return

    const fetchRelatedModels = async () => {
      setLoadingRelated(true)
      try {
        const response = await api.get('/models')
        const allModels = response.data.data?.models || []
        const related = allModels.filter((m: any) => 
          m.id !== model.id && 
          (m.provider_name === model.provider_name || m.category === model.category)
        ).slice(0, 4)
        setRelatedModels(related)
      } catch (error) {
        console.error('Failed to fetch related models:', error)
      } finally {
        setLoadingRelated(false)
      }
    }

    fetchRelatedModels()
  }, [model])

  const fetchApiKeys = async () => {
    try {
      const response = await api.get('/user/api-keys')
      const keys = response.data.data?.api_keys || []
      setApiKeys(keys)
      if (keys.length > 0) {
        setSelectedApiKey(keys[0].api_key)
      }
    } catch (error) {
      console.error('Failed to fetch API keys:', error)
      toast.error('Failed to load API keys')
    }
  }

  const testModel = async () => {
    if (!testMessage.trim() || !selectedApiKey) {
      toast.error('Please enter a message and ensure you have an API key')
      return
    }

    setIsTesting(true)
    setTestResponse('')

    try {
      const response = await api.post('/chat/completions', {
        model: model?.name,
        messages: [
          {
            role: 'user',
            content: testMessage
          }
        ],
        max_tokens: model?.max_tokens || 1000
      }, {
        headers: {
          'X-API-Key': selectedApiKey
        }
      })

      const content = response.data.choices?.[0]?.message?.content || 'No response content'
      setTestResponse(content)
      toast.success('Model test successful!')
    } catch (error: any) {
      console.error('Model test failed:', error)
      const errorMessage = error.response?.data?.error?.message || 'Failed to test model'
      setTestResponse(`Error: ${errorMessage}`)
      toast.error(`Test failed: ${errorMessage}`)
    } finally {
      setIsTesting(false)
    }
  }

  const copyToClipboard = async (text: string) => {
    try {
      await navigator.clipboard.writeText(text)
      setCopied(true)
      toast.success('Copied to clipboard!')
      setTimeout(() => setCopied(false), 2000)
    } catch (err) {
      toast.error('Failed to copy to clipboard')
    }
  }

  const handleTryButtonClick = () => {
    if (apiKeys.length === 0) {
      toast.info('You need an API key to test models. Please create one first.')
      router.push('/api-keys')
      return
    }
    setTryModalOpen(true)
  }

   if (isLoading) {
    return (
      <ProtectedRoute>
        <div className="min-h-screen bg-gray-50">
          <div className="p-6">
            <div className="max-w-7xl mx-auto">
              <div className="animate-pulse">
                <div className="h-8 w-48 bg-gray-200 rounded mb-6"></div>
                <div className="h-96 bg-gray-200 rounded mb-6"></div>
                <div className="h-64 bg-gray-200 rounded"></div>
              </div>
            </div>
          </div>
        </div>
      </ProtectedRoute>
    )
  }

  if (error || !model) {
    return (
      <ProtectedRoute>
        <div className="min-h-screen bg-gray-50">
          <div className="p-6">
            <div className="max-w-7xl mx-auto">
              <Button variant="outline" onClick={() => router.back()} className="mb-6">
                <ArrowLeft className="mr-2 h-4 w-4" />
                Back to Models
              </Button>
              <Card className="border-red-200 bg-red-50">
                <CardContent className="pt-6">
                  <div className="text-center">
                    <XCircle className="h-12 w-12 text-red-500 mx-auto mb-4" />
                    <h3 className="text-lg font-medium text-red-800">Model Not Found</h3>
                    <p className="text-red-600 mt-2">
                      {error?.message || 'The requested model could not be found.'}
                    </p>
                    <Button onClick={() => router.push('/models')} className="mt-4">
                      Browse All Models
                    </Button>
                  </div>
                </CardContent>
              </Card>
            </div>
          </div>
        </div>
      </ProtectedRoute>
    )
  }

  const formatPrice = (price: number) => {
    if (price === 0) return 'Free'
    return `$${price.toFixed(6)} per 1K tokens`
  }

  const capabilities = model.capabilities || {}
  const capabilityEntries = Object.entries(capabilities)

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="sticky top-0 z-50 border-b bg-white">
        <div className="flex h-16 items-center justify-between px-6">
          <div className="flex items-center gap-3">
            <Button variant="ghost" size="sm" onClick={() => router.push('/models')}>
              <ArrowLeft className="h-4 w-4 mr-2" />
              Back to Models
            </Button>
            <div className="h-6 w-px bg-gray-200"></div>
            <Zap className="h-5 w-5 text-primary" />
            <span className="text-sm font-medium text-gray-600">MassRouter Portal</span>
          </div>
          <Button onClick={() => router.push('/dashboard')}>
            Go to Dashboard
          </Button>
        </div>
      </header>

      <main className="p-6">
        <div className="max-w-7xl mx-auto">
          {/* Model Header */}
          <div className="mb-8">
            <div className="flex items-center justify-between">
              <div>
                <div className="flex items-center gap-3 mb-2">
                  <h1 className="text-3xl font-bold tracking-tight">{model.name}</h1>
                  <Badge variant={model.is_free ? "default" : "secondary"}>
                    {model.is_free ? 'Free' : 'Paid'}
                  </Badge>
                </div>
                <div className="flex items-center gap-4 text-muted-foreground">
                  <div className="flex items-center gap-1">
                    <Globe className="h-4 w-4" />
                    <span className="font-medium">{model.provider_name}</span>
                  </div>
                  {model.category && (
                    <div className="flex items-center gap-1">
                      <Database className="h-4 w-4" />
                      <span>{model.category}</span>
                    </div>
                  )}
                  {model.pricing_tier && (
                    <div className="flex items-center gap-1">
                      <DollarSign className="h-4 w-4" />
                      <span>{model.pricing_tier}</span>
                    </div>
                  )}
                </div>
              </div>
              <div className="flex gap-3">
                <Button size="lg" className="bg-gradient-primary hover:opacity-90" onClick={handleTryButtonClick}>
                  <Code className="mr-2 h-5 w-5" />
                  Try This Model
                </Button>
                <Button size="lg" variant={isComparing ? "destructive" : "outline"} onClick={() => isComparing ? removeModel(modelId) : addModel(modelId)}>
                  {isComparing ? (
                    <>
                      <XCircle className="mr-2 h-5 w-5" />
                      Remove from Compare ({count})
                    </>
                  ) : (
                    <>
                      <BarChart className="mr-2 h-5 w-5" />
                      Add to Compare ({count})
                    </>
                  )}
                </Button>
              </div>
            </div>
            
            {model.description && (
              <p className="mt-4 text-lg text-gray-700 max-w-3xl">
                {model.description}
              </p>
            )}
          </div>

          <div className="grid gap-6 lg:grid-cols-3">
            {/* Left Column - Model Details */}
            <div className="lg:col-span-2 space-y-6">
              {/* Pricing Card */}
              <Card>
                <CardHeader>
                  <CardTitle className="flex items-center gap-2">
                    <DollarSign className="h-5 w-5" />
                    Pricing
                  </CardTitle>
                  <CardDescription>Cost per 1,000 tokens (prompt + completion)</CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="grid gap-6 md:grid-cols-2">
                    <div className="space-y-2">
                      <div className="flex items-center justify-between">
                        <span className="text-sm font-medium text-gray-600">Input (Prompt)</span>
                        <Badge variant="outline" className={model.input_price === 0 ? 'bg-green-50 text-green-700' : ''}>
                          {model.input_price === 0 ? 'Free' : 'Paid'}
                        </Badge>
                      </div>
                      <div className="text-3xl font-bold">
                        {model.input_price === 0 ? 'Free' : `$${model.input_price.toFixed(6)}`}
                      </div>
                      <p className="text-sm text-gray-500">
                        Per 1,000 input tokens
                      </p>
                    </div>
                    
                    <div className="space-y-2">
                      <div className="flex items-center justify-between">
                        <span className="text-sm font-medium text-gray-600">Output (Completion)</span>
                        <Badge variant="outline" className={model.output_price === 0 ? 'bg-green-50 text-green-700' : ''}>
                          {model.output_price === 0 ? 'Free' : 'Paid'}
                        </Badge>
                      </div>
                      <div className="text-3xl font-bold">
                        {model.output_price === 0 ? 'Free' : `$${model.output_price.toFixed(6)}`}
                      </div>
                      <p className="text-sm text-gray-500">
                        Per 1,000 output tokens
                      </p>
                    </div>
                  </div>
                  
                  <div className="mt-6 p-4 bg-blue-50 rounded-lg">
                    <p className="text-sm text-blue-700">
                      💡 <strong>Example:</strong> A request with 500 input tokens and 300 output tokens would cost:
                      <br />
                      <code className="mt-1 inline-block bg-white px-2 py-1 rounded border">
                        (500 × ${model.input_price.toFixed(6)} / 1000) + (300 × ${model.output_price.toFixed(6)} / 1000) = 
                        ${((500 * model.input_price + 300 * model.output_price) / 1000).toFixed(6)}
                      </code>
                    </p>
                  </div>
                </CardContent>
              </Card>

              {/* Specifications */}
              <Card>
                <CardHeader>
                  <CardTitle className="flex items-center gap-2">
                    <Cpu className="h-5 w-5" />
                    Specifications
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="grid gap-4 md:grid-cols-2">
                    <div className="space-y-4">
                      <div>
                        <p className="text-sm font-medium text-gray-600 mb-1">Context Length</p>
                        <div className="flex items-center gap-2">
                          <Database className="h-4 w-4 text-primary" />
                          <span className="text-xl font-bold">
                            {model.context_length ? `${model.context_length.toLocaleString()} tokens` : 'Not specified'}
                          </span>
                        </div>
                        <p className="text-sm text-gray-500 mt-1">
                          Maximum context window size
                        </p>
                      </div>
                      
                      <div>
                        <p className="text-sm font-medium text-gray-600 mb-1">Max Tokens</p>
                        <div className="flex items-center gap-2">
                          <BarChart className="h-4 w-4 text-primary" />
                          <span className="text-xl font-bold">
                            {model.max_tokens ? `${model.max_tokens.toLocaleString()} tokens` : 'Not specified'}
                          </span>
                        </div>
                        <p className="text-sm text-gray-500 mt-1">
                          Maximum generation length
                        </p>
                      </div>
                    </div>
                    
                    <div>
                      <p className="text-sm font-medium text-gray-600 mb-3">Capabilities</p>
                      {capabilityEntries.length > 0 ? (
                        <div className="space-y-2">
                          {capabilityEntries.map(([key, value]) => (
                            <div key={key} className="flex items-center justify-between">
                              <span className="text-sm capitalize">{key.replace(/_/g, ' ')}</span>
                              {value === true ? (
                                <CheckCircle className="h-4 w-4 text-green-500" />
                              ) : value === false ? (
                                <XCircle className="h-4 w-4 text-red-500" />
                              ) : (
                                <span className="text-sm text-gray-600">{String(value)}</span>
                              )}
                            </div>
                          ))}
                        </div>
                      ) : (
                        <p className="text-sm text-gray-500">No specific capabilities listed</p>
                      )}
                    </div>
                  </div>
                </CardContent>
              </Card>

              {/* API Usage */}
              <Card>
                <CardHeader>
                  <CardTitle className="flex items-center gap-2">
                    <Code className="h-5 w-5" />
                    API Usage
                  </CardTitle>
                  <CardDescription>How to use this model with the MassRouter API</CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="space-y-4">
                    <div className="flex space-x-2 border-b">
                      <Button
                        variant={selectedTab === 'curl' ? 'default' : 'ghost'}
                        size="sm"
                        onClick={() => setSelectedTab('curl')}
                        className="rounded-b-none"
                      >
                        cURL
                      </Button>
                      <Button
                        variant={selectedTab === 'python' ? 'default' : 'ghost'}
                        size="sm"
                        onClick={() => setSelectedTab('python')}
                        className="rounded-b-none"
                      >
                        Python
                      </Button>
                      <Button
                        variant={selectedTab === 'javascript' ? 'default' : 'ghost'}
                        size="sm"
                        onClick={() => setSelectedTab('javascript')}
                        className="rounded-b-none"
                      >
                        JavaScript
                      </Button>
                    </div>
                    
                    {selectedTab === 'curl' && (
                      <pre className="bg-gray-900 text-gray-100 p-4 rounded-lg overflow-x-auto text-sm">
{`curl -X POST "http://localhost:8089/api/v1/chat/completions" \\
  -H "Authorization: Bearer YOUR_API_KEY" \\
  -H "Content-Type: application/json" \\
  -d '{
    "model": "${model.name}",
    "messages": [
      {"role": "user", "content": "Hello, how are you?"}
    ],
    "max_tokens": ${model.max_tokens || 1000}
  }'`}
                      </pre>
                    )}
                    
                    {selectedTab === 'python' && (
                      <pre className="bg-gray-900 text-gray-100 p-4 rounded-lg overflow-x-auto text-sm">
{`import requests

url = "http://localhost:8089/api/v1/chat/completions"
headers = {
    "Authorization": "Bearer YOUR_API_KEY",
    "Content-Type": "application/json"
}
data = {
    "model": "${model.name}",
    "messages": [
        {"role": "user", "content": "Hello, how are you?"}
    ],
    "max_tokens": ${model.max_tokens || 1000}
}

response = requests.post(url, headers=headers, json=data)
print(response.json())`}
                      </pre>
                    )}
                    
                    {selectedTab === 'javascript' && (
                      <pre className="bg-gray-900 text-gray-100 p-4 rounded-lg overflow-x-auto text-sm">
{`const response = await fetch('http://localhost:8089/api/v1/chat/completions', {
  method: 'POST',
  headers: {
    'Authorization': 'Bearer YOUR_API_KEY',
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    model: '${model.name}',
    messages: [
      { role: 'user', content: 'Hello, how are you?' }
    ],
    max_tokens: ${model.max_tokens || 1000}
  })
});

const data = await response.json();
console.log(data);`}
                      </pre>
                    )}
                  </div>
                </CardContent>
              </Card>
            </div>

            {/* Right Column - Sidebar */}
            <div className="space-y-6">
              {/* Provider Info */}
              <Card>
                <CardHeader>
                  <CardTitle className="text-lg">Provider</CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="space-y-4">
                    <div className="flex items-center gap-3">
                      <div className="h-10 w-10 rounded-full bg-gradient-primary flex items-center justify-center text-white font-bold">
                        {model.provider_name?.charAt(0).toUpperCase()}
                      </div>
                      <div>
                        <h4 className="font-semibold">{model.provider_name}</h4>
                        <p className="text-sm text-gray-500">AI Model Provider</p>
                      </div>
                    </div>
                    
                    <div className="space-y-2">
                      <div className="flex items-center justify-between">
                        <span className="text-sm text-gray-600">Status</span>
                        <Badge variant="outline" className="bg-green-50 text-green-700">
                          {model.is_active ? 'Active' : 'Inactive'}
                        </Badge>
                      </div>
                      <div className="flex items-center justify-between">
                        <span className="text-sm text-gray-600">Created</span>
                        <span className="text-sm">
                          {new Date(model.created_at).toLocaleDateString()}
                        </span>
                      </div>
                      <div className="flex items-center justify-between">
                        <span className="text-sm text-gray-600">Updated</span>
                        <span className="text-sm">
                          {new Date(model.updated_at).toLocaleDateString()}
                        </span>
                      </div>
                    </div>
                    
                    <Button variant="outline" className="w-full">
                      View All Models from {model.provider_name}
                      <ChevronRight className="ml-2 h-4 w-4" />
                    </Button>
                  </div>
                </CardContent>
              </Card>

              {/* Quick Actions */}
              <Card>
                <CardHeader>
                  <CardTitle className="text-lg">Quick Actions</CardTitle>
                </CardHeader>
                <CardContent className="space-y-3">
                  <Button className="w-full justify-start" variant="outline">
                    <Code className="mr-2 h-4 w-4" />
                    Generate API Key
                  </Button>
                  <Button className="w-full justify-start" variant="outline">
                    <BarChart className="mr-2 h-4 w-4" />
                    View Usage Analytics
                  </Button>
                  <Button className="w-full justify-start" variant="outline">
                    <DollarSign className="mr-2 h-4 w-4" />
                    Estimate Cost
                  </Button>
                  <Button className="w-full justify-start" variant="outline" onClick={() => {
                    if (isComparing) {
                      removeModel(modelId)
                      toast.success('Model removed from comparison')
                    } else {
                      addModel(modelId)
                      toast.success('Model added to comparison')
                    }
                  }}>
                    <BarChart className="mr-2 h-4 w-4" />
                    {isComparing ? 'Remove from Comparison' : 'Add to Comparison'} ({count})
                  </Button>
                </CardContent>
              </Card>

              {/* Pricing Comparison */}
              <Card>
                <CardHeader>
                  <CardTitle className="text-lg">Price Comparison</CardTitle>
                  <CardDescription>vs similar models</CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="space-y-3">
                    <div className="flex items-center justify-between p-3 bg-blue-50 rounded-lg">
                      <div>
                        <p className="font-medium text-sm">This Model</p>
                        <p className="text-xs text-gray-500">Input: ${model.input_price.toFixed(6)}</p>
                      </div>
                      <Badge variant="default">Current</Badge>
                    </div>
                    
                    <div className="text-sm text-gray-600 p-3 border rounded-lg">
                      <p className="font-medium mb-1">GPT-4o</p>
                      <p className="text-xs text-gray-500">Input: $0.005000</p>
                    </div>
                    
                    <div className="text-sm text-gray-600 p-3 border rounded-lg">
                      <p className="font-medium mb-1">Claude 3 Sonnet</p>
                      <p className="text-xs text-gray-500">Input: $0.003000</p>
                    </div>
                    
                    <Button variant="ghost" className="w-full text-primary">
                      Compare with 5 more models
                      <ChevronRight className="ml-2 h-4 w-4" />
                    </Button>
                  </div>
                </CardContent>
              </Card>
            </div>
          </div>

          {/* Call to Action */}
          <Card className="mt-8 border-primary bg-gradient-to-r from-primary/5 to-primary/10">
            <CardContent className="pt-6">
              <div className="flex flex-col md:flex-row items-center justify-between">
                <div>
                  <h3 className="text-xl font-bold">Ready to start using {model.name}?</h3>
                  <p className="text-gray-600 mt-1">
                    Get started with a free API key and begin integrating AI into your applications.
                  </p>
                </div>
                <div className="flex gap-3 mt-4 md:mt-0">
                  <Button size="lg" className="bg-gradient-primary hover:opacity-90">
                    Get API Key
                  </Button>
                  <Button size="lg" variant="outline">
                    View Documentation
                  </Button>
                </div>
              </div>
            </CardContent>
              </Card>

              {/* Related Models */}
              {relatedModels.length > 0 && (
                <Card>
                  <CardHeader>
                    <CardTitle className="flex items-center gap-2">
                      <BarChart className="h-5 w-5" />
                      Related Models
                    </CardTitle>
                    <CardDescription>Other models you might be interested in</CardDescription>
                  </CardHeader>
                  <CardContent>
                    <div className="grid gap-4 md:grid-cols-2">
                      {relatedModels.map((related) => (
                        <div key={related.id} className="border rounded-lg p-4 hover:bg-gray-50 transition-colors">
                          <div className="flex items-center justify-between mb-2">
                            <h4 className="font-semibold">{related.name}</h4>
                            <Badge variant="outline">{related.provider_name}</Badge>
                          </div>
                          <p className="text-sm text-gray-600 mb-3 line-clamp-2">{related.description || 'No description'}</p>
                          <div className="flex items-center justify-between text-sm">
                            <span className="text-gray-500">
                              Input: ${related.input_price?.toFixed(6) || '0.000000'}
                            </span>
                            <Button variant="ghost" size="sm" asChild>
                              <Link href={`/models/${related.id}`}>
                                View Details
                                <ChevronRight className="ml-1 h-3 w-3" />
                              </Link>
                            </Button>
                          </div>
                        </div>
                      ))}
                    </div>
                    {loadingRelated && (
                      <div className="text-center py-4">
                        <Loader2 className="h-6 w-6 animate-spin mx-auto" />
                      </div>
                    )}
                  </CardContent>
                </Card>
              )}
            </div>

        {/* Try Model Dialog */}
        <Dialog open={tryModalOpen} onOpenChange={setTryModalOpen}>
          <DialogContent className="max-w-4xl max-h-[90vh] overflow-y-auto">
            <DialogHeader>
              <DialogTitle>Test {model?.name}</DialogTitle>
              <DialogDescription>
                Send a test message to {model?.name} and see the response in real-time.
              </DialogDescription>
            </DialogHeader>
            
            <div className="grid gap-6 md:grid-cols-2">
              {/* Left Column - Input */}
              <div className="space-y-4">
                <div>
                  <Label htmlFor="api-key">API Key</Label>
                  <div className="flex gap-2 mt-1">
                    <select
                      id="api-key"
                      className="flex-1 border rounded-lg px-3 py-2 text-sm"
                      value={selectedApiKey}
                      onChange={(e) => setSelectedApiKey(e.target.value)}
                    >
                      {apiKeys.map((key) => (
                        <option key={key.id} value={key.api_key}>
                          {key.name} ({key.prefix}•••••••)
                        </option>
                      ))}
                    </select>
                    <Button
                      type="button"
                      variant="outline"
                      size="icon"
                      onClick={() => selectedApiKey && copyToClipboard(selectedApiKey)}
                      disabled={!selectedApiKey}
                    >
                      {copied ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}
                    </Button>
                  </div>
                  <p className="text-xs text-gray-500 mt-1">
                    Using key: {selectedApiKey ? `${selectedApiKey.substring(0, 8)}...` : 'None selected'}
                  </p>
                  <Button
                    variant="link"
                    className="text-xs p-0 h-auto"
                    onClick={() => router.push('/api-keys')}
                  >
                    Manage API Keys
                  </Button>
                </div>

                <div>
                  <Label htmlFor="test-message">Test Message</Label>
                  <textarea
                    id="test-message"
                    className="w-full h-40 border rounded-lg p-3 text-sm mt-1"
                    placeholder="Enter your message here..."
                    value={testMessage}
                    onChange={(e) => setTestMessage(e.target.value)}
                  />
                  <p className="text-xs text-gray-500 mt-1">
                    This will be sent as a user message to the model.
                  </p>
                </div>

                <Button
                  onClick={testModel}
                  disabled={isTesting || !testMessage.trim() || !selectedApiKey}
                  className="w-full"
                >
                  {isTesting ? (
                    <>
                      <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                      Testing...
                    </>
                  ) : (
                    <>
                      <Send className="mr-2 h-4 w-4" />
                      Send Test Message
                    </>
                  )}
                </Button>

                <div className="text-sm text-gray-600">
                  <p className="font-medium mb-1">Cost Estimate:</p>
                  <p className="text-xs">
                    Input: ${model?.input_price?.toFixed(6) || '0.000000'} per 1K tokens
                    <br />
                    Output: ${model?.output_price?.toFixed(6) || '0.000000'} per 1K tokens
                    <br />
                    <span className="text-gray-500">
                      Estimated cost for this test: ${((testMessage.length / 4) * (model?.input_price || 0) / 1000).toFixed(6)}
                    </span>
                  </p>
                </div>
              </div>

              {/* Right Column - Output */}
              <div className="space-y-4">
                <Label>Model Response</Label>
                <div className="h-64 border rounded-lg p-4 overflow-y-auto bg-gray-50">
                  {testResponse ? (
                    <div className="space-y-3">
                      <div className="flex items-start gap-3">
                        <div className="h-8 w-8 rounded-full bg-blue-100 flex items-center justify-center">
                          <Bot className="h-4 w-4 text-blue-600" />
                        </div>
                        <div className="flex-1">
                          <div className="flex items-center justify-between mb-1">
                            <span className="font-medium text-sm">{model?.name}</span>
                            <Button
                              variant="ghost"
                              size="sm"
                              onClick={() => copyToClipboard(testResponse)}
                            >
                              {copied ? <Check className="h-3 w-3" /> : <Copy className="h-3 w-3" />}
                            </Button>
                          </div>
                          <p className="text-sm text-gray-700 whitespace-pre-wrap">{testResponse}</p>
                        </div>
                      </div>
                    </div>
                  ) : (
                    <div className="h-full flex flex-col items-center justify-center text-gray-400">
                      <Bot className="h-12 w-12 mb-3" />
                      <p>Response will appear here...</p>
                      <p className="text-xs mt-1">Send a message to test the model</p>
                    </div>
                  )}
                </div>

                <div className="text-sm text-gray-600">
                  <p className="font-medium mb-1">Model Details:</p>
                  <div className="grid grid-cols-2 gap-2 text-xs">
                    <div>
                      <span className="text-gray-500">Provider:</span>
                      <span className="ml-2">{model?.provider_name}</span>
                    </div>
                    <div>
                      <span className="text-gray-500">Context:</span>
                      <span className="ml-2">{model?.context_length?.toLocaleString() || 'N/A'} tokens</span>
                    </div>
                    <div>
                      <span className="text-gray-500">Max Tokens:</span>
                      <span className="ml-2">{model?.max_tokens?.toLocaleString() || '1000'} tokens</span>
                    </div>
                    <div>
                      <span className="text-gray-500">Status:</span>
                      <span className="ml-2">
                        {model?.is_active ? (
                          <span className="text-green-600">Active</span>
                        ) : (
                          <span className="text-red-600">Inactive</span>
                        )}
                      </span>
                    </div>
                  </div>
                </div>

                <div className="text-xs text-gray-500 border-t pt-3">
                  <p className="font-medium mb-1">API Endpoint:</p>
                  <code className="block bg-gray-100 p-2 rounded text-xs overflow-x-auto">
                    POST http://localhost:8089/api/v1/chat/completions
                  </code>
                  <p className="mt-2">
                    This test uses the same endpoint as your production API calls.
                  </p>
                </div>
              </div>
            </div>

            <DialogFooter className="sm:justify-start">
              <Button variant="outline" onClick={() => setTryModalOpen(false)}>
                Close
              </Button>
              <Button
                variant="default"
                onClick={() => {
                  setTestMessage('')
                  setTestResponse('')
                }}
              >
                Clear
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </main>
    </div>
  )
}
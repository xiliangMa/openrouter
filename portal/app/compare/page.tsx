/* eslint-disable react/no-unescaped-entities */
'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import ProtectedRoute from '@/components/protected-route'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { useCompare } from '@/lib/hooks/useCompare'
import { useModelDetails } from '@/lib/hooks/useModels'
import { ArrowLeft, BarChart, Trash2, ExternalLink, DollarSign, Database, Cpu, CheckCircle, XCircle } from 'lucide-react'
import Link from 'next/link'
import { toast } from 'sonner'
import api from '@/lib/api'

interface ModelDetails {
  id: string
  name: string
  provider_name: string
  category: string
  input_price: number
  output_price: number
  context_length: number | null
  max_tokens: number | null
  is_free: boolean
  is_active: boolean
  capabilities: Record<string, any>
  pricing_tier: string
  description?: string
}

export default function ComparePage() {
  const router = useRouter()
  const { modelIds, models: compareModels, clearModels, removeModel, setModelsDetails } = useCompare()
  const [modelDetails, setModelDetails] = useState<ModelDetails[]>([])
  const [isLoading, setIsLoading] = useState(false)

  // Fetch details for each model
  useEffect(() => {
    if (modelIds.length === 0) return

    const fetchAllModelDetails = async () => {
      setIsLoading(true)
      try {
        const details = await Promise.all(
          modelIds.map(async (id) => {
            try {
              const response = await api.get(`/models/${id}`)
              return response.data.data as ModelDetails
            } catch (error) {
              console.error(`Failed to fetch model ${id}:`, error)
              return null
            }
          })
        )

        const validDetails = details.filter((d): d is ModelDetails => d !== null)
        setModelDetails(validDetails)
        
        // Update compare context with model details
        setModelsDetails(validDetails.map(model => ({
          id: model.id,
          name: model.name,
          provider: model.provider_name,
          category: model.category,
          inputPrice: model.input_price,
          outputPrice: model.output_price,
          contextLength: model.context_length || undefined,
          maxTokens: model.max_tokens || undefined
        })))
      } catch (error) {
        console.error('Failed to fetch model details:', error)
        toast.error('Failed to load model details')
      } finally {
        setIsLoading(false)
      }
    }

    fetchAllModelDetails()
  }, [modelIds, setModelsDetails])

  const handleClearAll = () => {
    clearModels()
    toast.success('Comparison cleared')
  }

  const handleRemoveModel = (modelId: string) => {
    removeModel(modelId)
    toast.success('Model removed from comparison')
  }

  if (modelIds.length === 0) {
    return (
      <ProtectedRoute>
        <div className="min-h-screen bg-gray-50">
          <div className="p-6">
            <div className="max-w-7xl mx-auto">
              <Button variant="outline" onClick={() => router.push('/models')} className="mb-6">
                <ArrowLeft className="mr-2 h-4 w-4" />
                Back to Models
              </Button>
              
              <Card className="text-center">
                <CardContent className="pt-12 pb-12">
                  <BarChart className="h-16 w-16 text-gray-300 mx-auto mb-4" />
                  <h2 className="text-2xl font-bold text-gray-700 mb-2">No Models to Compare</h2>
                  <p className="text-gray-500 mb-6">
                    Add models to comparison from the model details page to see them side by side.
                  </p>
                  <Button onClick={() => router.push('/models')}>
                    Browse Models
                  </Button>
                </CardContent>
              </Card>
            </div>
          </div>
        </div>
      </ProtectedRoute>
    )
  }

  if (isLoading) {
    return (
      <ProtectedRoute>
        <div className="min-h-screen bg-gray-50">
          <div className="p-6">
            <div className="max-w-7xl mx-auto">
              <Button variant="outline" onClick={() => router.push('/models')} className="mb-6">
                <ArrowLeft className="mr-2 h-4 w-4" />
                Back to Models
              </Button>
              <div className="animate-pulse">
                <div className="h-8 w-48 bg-gray-200 rounded mb-6"></div>
                <div className="h-64 bg-gray-200 rounded mb-6"></div>
                <div className="h-96 bg-gray-200 rounded"></div>
              </div>
            </div>
          </div>
        </div>
      </ProtectedRoute>
    )
  }

  const formatPrice = (price: number) => {
    if (price === 0) return 'Free'
    return `$${price.toFixed(6)}`
  }

  const capabilityKeys = Array.from(
    new Set(
      modelDetails.flatMap(model => 
        Object.keys(model.capabilities || {})
      )
    )
  )

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="sticky top-0 z-50 border-b bg-white">
        <div className="flex h-16 items-center justify-between px-6">
          <div className="flex items-center gap-3">
            <Button variant="ghost" size="sm" onClick={() => router.push('/models')}>
              <ArrowLeft className="h-4 w-4 mr-2" />
              Back to Models
            </Button>
            <div className="h-6 w-px bg-gray-200"></div>
            <BarChart className="h-5 w-5 text-primary" />
            <span className="text-sm font-medium text-gray-600">Model Comparison</span>
          </div>
          <div className="flex items-center gap-3">
            <Button variant="outline" onClick={handleClearAll}>
              <Trash2 className="mr-2 h-4 w-4" />
              Clear All
            </Button>
            <Button onClick={() => router.push('/models')}>
              Add More Models
            </Button>
          </div>
        </div>
      </header>

      <main className="p-6">
        <div className="max-w-7xl mx-auto">
          <div className="mb-8">
            <h1 className="text-3xl font-bold tracking-tight mb-2">Model Comparison</h1>
            <p className="text-gray-600">
              Compare {modelDetails.length} model{modelDetails.length !== 1 ? 's' : ''} side by side
            </p>
          </div>

          {/* Model Cards */}
          <div className="grid gap-6 mb-8 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
            {modelDetails.map((model) => (
              <Card key={model.id} className="relative">
                <Button
                  variant="ghost"
                  size="icon"
                  className="absolute right-2 top-2 h-8 w-8"
                  onClick={() => handleRemoveModel(model.id)}
                >
                  <Trash2 className="h-4 w-4" />
                </Button>
                <CardHeader className="pb-3">
                  <CardTitle className="text-lg">{model.name}</CardTitle>
                  <CardDescription>{model.provider_name}</CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="space-y-3">
                    <div className="flex items-center justify-between">
                      <span className="text-sm text-gray-600">Category</span>
                      <Badge variant="outline">{model.category}</Badge>
                    </div>
                    <div className="flex items-center justify-between">
                      <span className="text-sm text-gray-600">Pricing Tier</span>
                      <Badge variant={model.is_free ? "default" : "secondary"}>
                        {model.pricing_tier}
                      </Badge>
                    </div>
                    <div className="flex items-center justify-between">
                      <span className="text-sm text-gray-600">Input Price</span>
                      <span className="text-sm font-medium">{formatPrice(model.input_price)}</span>
                    </div>
                    <div className="flex items-center justify-between">
                      <span className="text-sm text-gray-600">Output Price</span>
                      <span className="text-sm font-medium">{formatPrice(model.output_price)}</span>
                    </div>
                    <div className="pt-3 border-t">
                      <Link href={`/models/${model.id}`} className="w-full">
                        <Button variant="outline" className="w-full">
                          View Details
                          <ExternalLink className="ml-2 h-3 w-3" />
                        </Button>
                      </Link>
                    </div>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>

          {/* Comparison Table */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <BarChart className="h-5 w-5" />
                Detailed Comparison
              </CardTitle>
              <CardDescription>Side-by-side comparison of key features and capabilities</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="overflow-x-auto">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead className="w-[200px]">Feature</TableHead>
                      {modelDetails.map(model => (
                        <TableHead key={model.id} className="text-center">
                          {model.name}
                        </TableHead>
                      ))}
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {/* Provider */}
                    <TableRow>
                      <TableCell className="font-medium">Provider</TableCell>
                      {modelDetails.map(model => (
                        <TableCell key={model.id} className="text-center">
                          {model.provider_name}
                        </TableCell>
                      ))}
                    </TableRow>

                    {/* Category */}
                    <TableRow>
                      <TableCell className="font-medium">Category</TableCell>
                      {modelDetails.map(model => (
                        <TableCell key={model.id} className="text-center">
                          <Badge variant="outline">{model.category}</Badge>
                        </TableCell>
                      ))}
                    </TableRow>

                    {/* Pricing Tier */}
                    <TableRow>
                      <TableCell className="font-medium">Pricing Tier</TableCell>
                      {modelDetails.map(model => (
                        <TableCell key={model.id} className="text-center">
                          <Badge variant={model.is_free ? "default" : "secondary"}>
                            {model.pricing_tier}
                          </Badge>
                        </TableCell>
                      ))}
                    </TableRow>

                    {/* Input Price */}
                    <TableRow>
                      <TableCell className="font-medium">Input Price (per 1K tokens)</TableCell>
                      {modelDetails.map(model => (
                        <TableCell key={model.id} className="text-center">
                          <div className="flex items-center justify-center gap-1">
                            <DollarSign className="h-3 w-3" />
                            {formatPrice(model.input_price)}
                          </div>
                        </TableCell>
                      ))}
                    </TableRow>

                    {/* Output Price */}
                    <TableRow>
                      <TableCell className="font-medium">Output Price (per 1K tokens)</TableCell>
                      {modelDetails.map(model => (
                        <TableCell key={model.id} className="text-center">
                          <div className="flex items-center justify-center gap-1">
                            <DollarSign className="h-3 w-3" />
                            {formatPrice(model.output_price)}
                          </div>
                        </TableCell>
                      ))}
                    </TableRow>

                    {/* Context Length */}
                    <TableRow>
                      <TableCell className="font-medium">Context Length</TableCell>
                      {modelDetails.map(model => (
                        <TableCell key={model.id} className="text-center">
                          <div className="flex items-center justify-center gap-1">
                            <Database className="h-3 w-3" />
                            {model.context_length ? `${model.context_length.toLocaleString()} tokens` : 'N/A'}
                          </div>
                        </TableCell>
                      ))}
                    </TableRow>

                    {/* Max Tokens */}
                    <TableRow>
                      <TableCell className="font-medium">Max Tokens</TableCell>
                      {modelDetails.map(model => (
                        <TableCell key={model.id} className="text-center">
                          <div className="flex items-center justify-center gap-1">
                            <Cpu className="h-3 w-3" />
                            {model.max_tokens ? `${model.max_tokens.toLocaleString()} tokens` : 'N/A'}
                          </div>
                        </TableCell>
                      ))}
                    </TableRow>

                    {/* Status */}
                    <TableRow>
                      <TableCell className="font-medium">Status</TableCell>
                      {modelDetails.map(model => (
                        <TableCell key={model.id} className="text-center">
                          <Badge variant={model.is_active ? "default" : "destructive"}>
                            {model.is_active ? 'Active' : 'Inactive'}
                          </Badge>
                        </TableCell>
                      ))}
                    </TableRow>

                    {/* Capabilities */}
                    {capabilityKeys.map(capability => (
                      <TableRow key={capability}>
                        <TableCell className="font-medium capitalize">{capability.replace(/_/g, ' ')}</TableCell>
                        {modelDetails.map(model => {
                          const value = model.capabilities?.[capability]
                          return (
                            <TableCell key={model.id} className="text-center">
                              {value === true ? (
                                <CheckCircle className="h-4 w-4 text-green-500 mx-auto" />
                              ) : value === false ? (
                                <XCircle className="h-4 w-4 text-red-500 mx-auto" />
                              ) : (
                                <span className="text-sm">{String(value)}</span>
                              )}
                            </TableCell>
                          )
                        })}
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
            </CardContent>
          </Card>

          {/* Actions */}
          <div className="mt-8 flex justify-center gap-4">
            <Button variant="outline" onClick={handleClearAll} size="lg">
              <Trash2 className="mr-2 h-5 w-5" />
              Clear Comparison
            </Button>
            <Button onClick={() => router.push('/models')} size="lg">
              Add More Models
            </Button>
            <Button variant="default" onClick={() => router.push('/dashboard')} size="lg">
              Go to Dashboard
            </Button>
          </div>
        </div>
      </main>
    </div>
  )
}
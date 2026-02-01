import { useState, useEffect, useCallback } from 'react'

const COMPARE_STORAGE_KEY = 'massrouter_compare_models'

export interface CompareModel {
  id: string
  name: string
  provider?: string
  category?: string
  inputPrice?: number
  outputPrice?: number
  contextLength?: number
  maxTokens?: number
}

export function useCompare() {
  const [modelIds, setModelIds] = useState<string[]>([])
  const [models, setModels] = useState<CompareModel[]>([])

  // Load from localStorage on mount
  useEffect(() => {
    const stored = localStorage.getItem(COMPARE_STORAGE_KEY)
    if (stored) {
      try {
        const ids = JSON.parse(stored)
        if (Array.isArray(ids)) {
          setModelIds(ids)
        }
      } catch (err) {
        console.error('Failed to parse compare models from localStorage:', err)
      }
    }
  }, [])

  // Save to localStorage whenever modelIds change
  useEffect(() => {
    localStorage.setItem(COMPARE_STORAGE_KEY, JSON.stringify(modelIds))
  }, [modelIds])

  const addModel = useCallback((modelId: string) => {
    setModelIds(prev => {
      if (prev.includes(modelId)) {
        return prev
      }
      // Limit to 4 models for comparison
      if (prev.length >= 4) {
        return [...prev.slice(1), modelId]
      }
      return [...prev, modelId]
    })
  }, [])

  const removeModel = useCallback((modelId: string) => {
    setModelIds(prev => prev.filter(id => id !== modelId))
  }, [])

  const clearModels = useCallback(() => {
    setModelIds([])
  }, [])

  const setModelsDetails = useCallback((details: CompareModel[]) => {
    setModels(details)
  }, [])

  const isInCompare = useCallback((modelId: string) => {
    return modelIds.includes(modelId)
  }, [modelIds])

  return {
    modelIds,
    models,
    addModel,
    removeModel,
    clearModels,
    setModelsDetails,
    isInCompare,
    count: modelIds.length
  }
}
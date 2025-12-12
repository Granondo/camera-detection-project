'use client'

import { SimpleGrid } from '@chakra-ui/react'
import React from 'react'

interface GridSkeletonProps {
  count?: number
  skeletonCard: React.ElementType
  columns?: Record<string, unknown>
}

export function GridSkeleton({
  count = 6,
  skeletonCard: SkeletonCard,
  columns = { base: 1, sm: 2, lg: 3 },
}: GridSkeletonProps) {
  return (
    <SimpleGrid columns={columns} gap={6}>
      {Array.from({ length: count }).map((_, i) => (
        <SkeletonCard key={i} />
      ))}
    </SimpleGrid>
  )
}

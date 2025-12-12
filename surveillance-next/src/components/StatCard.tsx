'use client'

import { Box, Text } from '@chakra-ui/react'
import { ReactNode } from 'react'

interface StatCardProps {
  label: string
  children: ReactNode
}

export function StatCard({ label, children }: StatCardProps) {
  return (
    <Box bg="white" p={4} borderRadius="md" shadow="sm">
      <Text fontSize="sm" color="gray.500" mb={1}>
        {label}
      </Text>
      <Text fontWeight="medium">{children}</Text>
    </Box>
  )
}

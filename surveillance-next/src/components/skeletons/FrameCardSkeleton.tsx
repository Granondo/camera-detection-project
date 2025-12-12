'use client'

import { Box, Skeleton } from '@chakra-ui/react'

export function FrameCardSkeleton() {
  return (
    <Box overflow="hidden" borderWidth="1px" borderRadius="md">
      <Skeleton height="180px" width="100%" />
      <Box p={3}>
        <Skeleton height="15px" width="80%" mb={2} />
        <Skeleton height="10px" width="60%" />
      </Box>
    </Box>
  )
}


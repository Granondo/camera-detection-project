'use client'

import { Box, Skeleton, SkeletonText } from '@chakra-ui/react'

export function EventCardSkeleton() {
  return (
    <Box p={4} borderWidth="1px" borderRadius="md">
      <Skeleton height="20px" width="70%" mb={3} />
      <SkeletonText noOfLines={3} gap={3} />
      <Skeleton height="10px" width="50%" mt={4} />
    </Box>
  )
}

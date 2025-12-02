'use client'

import { Box, Card, Badge, Text, Image } from '@chakra-ui/react'
import NextLink from 'next/link'
import { format } from 'date-fns'
import type { Frame } from '@/lib/api'
import { getImageUrl } from '@/lib/api'

interface FrameCardProps {
  frame: Frame
}

export function FrameCard({ frame }: FrameCardProps) {
  // Защита от невалидной даты
  let formattedTime = 'Unknown time'
  try {
    const date = new Date(frame.timestamp)
    if (!isNaN(date.getTime())) {
      formattedTime = format(date, 'd MMM yyyy, HH:mm:ss')
    }
  } catch {
    formattedTime = 'Invalid date'
  }

  const detectionsCount = frame.detections?.length || 0

  return (
    <NextLink href={`/frames/${frame.id}`} style={{ textDecoration: 'none' }}>
      <Card.Root
        overflow="hidden"
        cursor="pointer"
        transition="all 0.2s"
        _hover={{ transform: 'translateY(-2px)', shadow: 'lg' }}
      >
        <Box position="relative">
          <Image
            src={getImageUrl(frame.image_path)}
            alt={`Frame ${frame.id}`}
            objectFit="cover"
            h="180px"
            w="100%"
          />
          {detectionsCount > 0 && (
            <Badge
              position="absolute"
              top={2}
              right={2}
              colorPalette="purple"
            >
              {detectionsCount} {detectionsCount === 1 ? 'detection' : 'detections'}
            </Badge>
          )}
        </Box>
        
        <Card.Body p={3}>
          <Text fontSize="sm" color="gray.600">
            Camera {frame.camera_id}
          </Text>
          <Text fontSize="xs" color="gray.400">
            {formattedTime}
          </Text>
        </Card.Body>
      </Card.Root>
    </NextLink>
  )
}
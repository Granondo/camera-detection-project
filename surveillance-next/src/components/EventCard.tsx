import { Box, Badge, Flex, Heading, Text, Link } from '@chakra-ui/react'
import NextLink from 'next/link'
import { format } from 'date-fns'
import type { Event } from '@/lib/api'

interface EventCardProps {
  event: Event
}

export function EventCard({ event }: EventCardProps) {
  const statusColor = event.status === 'active' ? 'green' : 'gray'
  const formattedDate = format(new Date(event.startTime), 'dd MMM yyyy, HH:mm')

  return (
    <Link as={NextLink} href={`/events/${event.id}`} _hover={{ textDecoration: 'none' }}>
      <Box
        bg="white"
        borderRadius="lg"
        shadow="sm"
        p={5}
        transition="all 0.2s"
        _hover={{ shadow: 'md', transform: 'translateY(-2px)' }}
      >
        <Flex justify="space-between" align="start" mb={3}>
          <Heading size="sm" color="gray.800">
            {event.type}
          </Heading>
          <Badge colorPalette={statusColor} variant="subtle">
            {event.status}
          </Badge>
        </Flex>

        <Flex direction="column" gap={1}>
          <Text fontSize="sm" color="gray.600">
            📷 Camera: {event.cameraId}
          </Text>
          <Text fontSize="sm" color="gray.600">
            🕐 {formattedDate}
          </Text>
          <Text fontSize="sm" color="gray.600">
            🖼️ Frames: {event.frames.length}
          </Text>
        </Flex>
      </Box>
    </Link>
  )
}
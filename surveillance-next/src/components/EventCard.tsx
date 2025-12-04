'use client'

import { Box, Card, Badge, Text, HStack, Link } from '@chakra-ui/react'
import NextLink from 'next/link'
import { format } from 'date-fns'
import type { Event } from '@/lib/api'

interface EventCardProps {
  event: Event
}

export function EventCard({ event }: EventCardProps) {
  const formattedTime = format(new Date(event.timestamp), 'd MMM yyyy, HH:mm')

  const severityColor = {
    low: 'green',
    medium: 'yellow',
    high: 'orange',
    critical: 'red',
  }[event.severity] || 'gray'

  return (
    <NextLink href={`/events/${event.id}`} style={{ textDecoration: 'none' }}>
      <Card.Root
        p={4}
        cursor="pointer"
        transition="all 0.2s"
        _hover={{ transform: 'translateY(-2px)', shadow: 'lg' }}
      >
        <Card.Body p={0}>
          <HStack justify="space-between" mb={2}>
            <Text fontWeight="bold" fontSize="lg">
              {event.title}
            </Text>
            <Badge colorPalette={severityColor}>{event.severity}</Badge>
          </HStack>

          <Text color="gray.600" fontSize="sm" mb={2}>
            {event.event_type}
          </Text>

          <Text color="gray.500" fontSize="sm" lineClamp={2} mb={3}>
            {event.message}
          </Text>

          <HStack justify="space-between" fontSize="xs" color="gray.400">
            <Text>Camera {event.camera_id}</Text>
            <Text>{formattedTime}</Text>
          </HStack>

          {event.frame_id && (
            <Box mt={3} pt={3} borderTop="1px solid" borderColor="gray.200">
              <NextLink
                href={`/frames/${event.frame_id}`}
                onClick={(e) => e.stopPropagation()}
              >
                <Text fontSize="xs" color="blue.500" fontWeight="medium" _hover={{ textDecoration: 'underline' }}>
                  🖼️ View frame #{event.frame_id}
                </Text>
              </NextLink>
            </Box>
          )}
        </Card.Body>
      </Card.Root>
    </NextLink>
  )
}
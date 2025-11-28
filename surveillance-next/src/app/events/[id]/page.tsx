import { Box, Container, Heading, Text, Badge, SimpleGrid, Flex } from '@chakra-ui/react'
import { format } from 'date-fns'
import { Header } from '@/components/Header'
import { FrameCard } from '@/components/FrameCard'
import { getEvent } from '@/lib/api'
import type { Metadata } from 'next'

interface EventDetailPageProps {
  params: Promise<{
    id: string
  }>
}

export async function generateMetadata({ params }: EventDetailPageProps): Promise<Metadata> {
  const { id } = await params
  const event = await getEvent(id)
  
  return {
    title: `${event.type} Event | Surveillance`,
    description: `Detection event from camera ${event.cameraId} at ${event.startTime}`,
  }
}

export default async function EventDetailPage({ params }: EventDetailPageProps) {
  const { id } = await params
  const event = await getEvent(id)
  
  const statusColor = event.status === 'active' ? 'green' : 'gray'
  const formattedStart = format(new Date(event.startTime), 'dd MMM yyyy, HH:mm:ss')
  const formattedEnd = event.endTime 
    ? format(new Date(event.endTime), 'dd MMM yyyy, HH:mm:ss')
    : 'Ongoing'

  return (
    <Box minH="100vh">
      <Header />

      <Container maxW="container.xl" py={8}>
        <Box mb={8}>
          <Flex align="center" gap={3} mb={3}>
            <Heading size="xl" color="gray.800">
              {event.type}
            </Heading>
            <Badge colorPalette={statusColor} variant="solid" fontSize="sm">
              {event.status}
            </Badge>
          </Flex>

          <SimpleGrid columns={{ base: 1, sm: 2, md: 4 }} gap={4} mb={6}>
            <Box bg="white" p={4} borderRadius="md" shadow="sm">
              <Text fontSize="sm" color="gray.500" mb={1}>Camera</Text>
              <Text fontWeight="medium">{event.cameraId}</Text>
            </Box>
            <Box bg="white" p={4} borderRadius="md" shadow="sm">
              <Text fontSize="sm" color="gray.500" mb={1}>Started</Text>
              <Text fontWeight="medium">{formattedStart}</Text>
            </Box>
            <Box bg="white" p={4} borderRadius="md" shadow="sm">
              <Text fontSize="sm" color="gray.500" mb={1}>Ended</Text>
              <Text fontWeight="medium">{formattedEnd}</Text>
            </Box>
            <Box bg="white" p={4} borderRadius="md" shadow="sm">
              <Text fontSize="sm" color="gray.500" mb={1}>Frames</Text>
              <Text fontWeight="medium">{event.frames.length}</Text>
            </Box>
          </SimpleGrid>
        </Box>

        <Box>
          <Heading size="lg" mb={4} color="gray.800">
            Event Frames
          </Heading>

          {event.frames.length > 0 ? (
            <SimpleGrid columns={{ base: 1, sm: 2, lg: 3, xl: 4 }} gap={6}>
              {event.frames.map((frame) => (
                <FrameCard key={frame.id} frame={frame} />
              ))}
            </SimpleGrid>
          ) : (
            <Box textAlign="center" py={12} bg="white" borderRadius="lg">
              <Text color="gray.500">No frames in this event</Text>
            </Box>
          )}
        </Box>
      </Container>
    </Box>
  )
}
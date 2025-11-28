import { Box, Container, Heading, Text, Badge, SimpleGrid, Flex } from '@chakra-ui/react'
import { format } from 'date-fns'
import { Header } from '@/components/Header'
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
  
  if (!event) {
    return {
      title: 'Event Not Found | Surveillance',
    }
  }
  
  return {
    title: `${event.title} | Surveillance`,
    description: `Detection event from camera ${event.camera_id} at ${event.timestamp}`,
  }
}

export default async function EventDetailPage({ params }: EventDetailPageProps) {
  const { id } = await params
  const event = await getEvent(id)
  
  if (!event) {
    return (
      <Box minH="100vh">
        <Header />
        <Container maxW="container.xl" py={8}>
          <Text>Event not found</Text>
        </Container>
      </Box>
    )
  }

  const severityColor = {
    low: 'green',
    medium: 'yellow',
    high: 'orange',
    critical: 'red',
  }[event.severity] || 'gray'

  const formattedTime = format(new Date(event.timestamp), 'dd MMM yyyy, HH:mm:ss')
  const formattedCreatedAt = format(new Date(event.created_at), 'dd MMM yyyy, HH:mm:ss')
  const formattedResolvedAt = event.resolved_at 
    ? format(new Date(event.resolved_at), 'dd MMM yyyy, HH:mm:ss')
    : 'Not resolved'

  return (
    <Box minH="100vh">
      <Header />
      
      <Container maxW="container.xl" py={8}>
        <Box mb={8}>
          <Flex align="center" gap={3} mb={3} flexWrap="wrap">
            <Heading size="xl" color="gray.800">
              {event.title}
            </Heading>
            <Badge colorPalette={severityColor} variant="solid" fontSize="sm">
              {event.severity}
            </Badge>
            <Badge colorPalette={event.resolved ? 'green' : 'orange'} variant="outline" fontSize="sm">
              {event.resolved ? 'Resolved' : 'Active'}
            </Badge>
          </Flex>
          
          <Text color="gray.600" fontSize="lg" mb={6}>
            {event.message}
          </Text>

          <SimpleGrid columns={{ base: 1, sm: 2, md: 4 }} gap={4} mb={6}>
            <Box bg="white" p={4} borderRadius="md" shadow="sm">
              <Text fontSize="sm" color="gray.500" mb={1}>Camera</Text>
              <Text fontWeight="medium">{event.camera_id}</Text>
            </Box>
            <Box bg="white" p={4} borderRadius="md" shadow="sm">
              <Text fontSize="sm" color="gray.500" mb={1}>Event Type</Text>
              <Text fontWeight="medium">{event.event_type}</Text>
            </Box>
            <Box bg="white" p={4} borderRadius="md" shadow="sm">
              <Text fontSize="sm" color="gray.500" mb={1}>Timestamp</Text>
              <Text fontWeight="medium">{formattedTime}</Text>
            </Box>
            <Box bg="white" p={4} borderRadius="md" shadow="sm">
              <Text fontSize="sm" color="gray.500" mb={1}>Created At</Text>
              <Text fontWeight="medium">{formattedCreatedAt}</Text>
            </Box>
          </SimpleGrid>

          <SimpleGrid columns={{ base: 1, sm: 2 }} gap={4}>
            <Box bg="white" p={4} borderRadius="md" shadow="sm">
              <Text fontSize="sm" color="gray.500" mb={1}>Resolved At</Text>
              <Text fontWeight="medium">{formattedResolvedAt}</Text>
            </Box>
            <Box bg="white" p={4} borderRadius="md" shadow="sm">
              <Text fontSize="sm" color="gray.500" mb={1}>Notified</Text>
              <Text fontWeight="medium">{event.notified ? 'Yes' : 'No'}</Text>
            </Box>
          </SimpleGrid>
        </Box>
      </Container>
    </Box>
  )
}
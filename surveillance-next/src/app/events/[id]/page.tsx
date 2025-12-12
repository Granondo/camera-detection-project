import { Suspense } from 'react'
import {
  Box,
  Container,
  Heading,
  Text,
  Badge,
  Grid,
  GridItem,
  Flex,
  Button,
  VStack,
  HStack,
} from '@chakra-ui/react'
import { format } from 'date-fns'
import { Header } from '@/components/Header'
import { getEvent } from '@/lib/api'
import type { Metadata } from 'next'
import { FrameViewer, FrameViewerSkeleton } from '@/components/FrameViewer'
import { MetadataViewer } from '@/components/MetadataViewer'
import { StatCard } from '@/components/StatCard'
import { NotFound } from '@/components/NotFound'

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
    description: `Details for detection event #${id}`,
  }
}

// Define CustomDivider outside the main component to avoid re-creation on render
const CustomDivider = () => (
  <Box pt={4} borderTop="1px solid" borderColor="gray.200" />
)

export default async function EventDetailPage({ params }: EventDetailPageProps) {
  const { id } = await params
  const event = await getEvent(id)

  if (!event) {
    return <NotFound title="Event Not Found" message={`The event with ID #${id} could not be found.`} />
  }

  const severityColor = {
    low: 'green',
    medium: 'yellow',
    high: 'orange',
    critical: 'red',
  }[event.severity] || 'gray'

  console.log('event', event)

  return (
    <Box minH="100vh">
      <Header />

      <Container maxW="6xl" py={8}>
        <Grid templateColumns={{ base: '1fr', lg: '2fr 1fr' }} gap={8}>
          {/* Left Column: Frame Viewer */}
          <GridItem>
            <VStack align="stretch">
              <Heading size="lg">Frame Analysis</Heading>
              {event.frame_id ? (
                <Suspense fallback={<FrameViewerSkeleton />}>
                  <FrameViewer frameId={event.frame_id} />
                </Suspense>
              ) : (
                <Box p={4} bg="gray.100" borderRadius="md" textAlign="center">
                  <Text color="gray.600">No frame associated with this event.</Text>
                </Box>
              )}
            </VStack>
          </GridItem>

          {/* Right Column: Event Details */}
          <GridItem>
            <VStack align="stretch" gap={5}>
              {/* Header */}
              <Box>
                <Flex align="flex-start" justify="space-between" gap={2}>
                  <Heading size="lg" color="gray.800" flex="1">
                    {event.title}
                  </Heading>
                  <Badge colorScheme={severityColor} variant="solid" fontSize="sm" mt={1}>
                    {event.severity}
                  </Badge>
                </Flex>
                <Text color="gray.500" mt={2}>
                  {event.message}
                </Text>
              </Box>

              {/* Action Buttons */}
              <HStack>
                <Button colorScheme="blue" disabled={event.resolved}>
                  Mark as Resolved
                </Button>
                <Button variant="outline">Acknowledge</Button>
              </HStack>
              
              <CustomDivider />

              {/* Main Details */}
              <VStack align="stretch" gap={3}>
                <StatCard label="Status">
                  <Badge colorScheme={event.resolved ? 'green' : 'orange'} variant="outline">
                    {event.resolved ? 'Resolved' : 'Active'}
                  </Badge>
                </StatCard>
                <StatCard label="Camera ID">{event.camera_id}</StatCard>
                <StatCard label="Event Type">{event.event_type}</StatCard>
                <StatCard label="Event Timestamp">
                  {format(new Date(event.timestamp), 'dd MMM yyyy, HH:mm:ss')}
                </StatCard>
                 <StatCard label="Notified">{event.notified ? 'Yes' : 'No'}</StatCard>
                {event.resolved_at && (
                  <StatCard label="Resolved At">
                    {format(new Date(event.resolved_at), 'dd MMM yyyy, HH:mm:ss')}
                  </StatCard>
                )}
              </VStack>
              
              <CustomDivider />

              {/* Metadata */}
              <MetadataViewer metadata={event.metadata} />
            </VStack>
          </GridItem>
        </Grid>
      </Container>
    </Box>
  )
}
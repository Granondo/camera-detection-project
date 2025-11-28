import { Box, Container, Heading, Text, SimpleGrid } from '@chakra-ui/react'
import { Header } from '@/components/Header'
import { EventCard } from '@/components/EventCard'
import { getEvents } from '@/lib/api'

interface EventsPageProps {
  searchParams: Promise<{
    page?: string
    type?: string
  }>
}

export const metadata = {
  title: 'Events | Surveillance',
  description: 'Browse all detection events from video surveillance',
}

export default async function EventsPage({ searchParams }: EventsPageProps) {
  const params = await searchParams
  const page = Number(params.page) || 1
  const type = params.type

  const { events, total } = await getEvents({
    page,
    perPage: 12,
    type,
  })

  return (
    <Box minH="100vh">
      <Header />

      <Container maxW="container.xl" py={8}>
        <Box mb={8}>
          <Heading size="xl" mb={2} color="gray.800">
            Detection Events
          </Heading>
          <Text color="gray.600">
            {total} event{total !== 1 ? 's' : ''} found
          </Text>
        </Box>

        {events.length > 0 ? (
          <SimpleGrid columns={{ base: 1, sm: 2, lg: 3 }} gap={6}>
            {events.map((event) => (
              <EventCard key={event.id} event={event} />
            ))}
          </SimpleGrid>
        ) : (
          <Box textAlign="center" py={12}>
            <Text fontSize="lg" color="gray.500">
              No events found
            </Text>
          </Box>
        )}
      </Container>
    </Box>
  )
}
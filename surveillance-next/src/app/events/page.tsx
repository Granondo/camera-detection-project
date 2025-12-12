import { Suspense } from 'react'
import { Box, Container, Heading, Text } from '@chakra-ui/react'
import { Header } from '@/components/Header'
import { GridSkeleton } from '@/components/skeletons/GridSkeleton'
import { EventCardSkeleton } from '@/components/skeletons/EventCardSkeleton'
import { EventsGrid } from './EventsGrid'
import { EventsToolbar } from './EventsToolbar'

interface EventsPageProps {
  searchParams: {
    page?: string
    type?: string
  }
}

export const metadata = {
  title: 'Events | Surveillance',
  description: 'Browse all detection events from video surveillance',
}

export default function EventsPage({ searchParams }: EventsPageProps) {
  const page = Number(searchParams?.page) || 1
  const type = searchParams?.type

  return (
    <Box minH="100vh">
      <Header />

      <Container maxW="container.xl" py={8}>
        <Box mb={4}>
          <Heading size="xl" mb={2} color="gray.800">
            Detection Events
          </Heading>
          <Text color="gray.600">
            Browse all detection events from video surveillance.
          </Text>
        </Box>

        <EventsToolbar />

        <Suspense
          key={`${page}-${type || ''}`}
          fallback={<GridSkeleton count={12} skeletonCard={EventCardSkeleton} />}
        >
          <EventsGrid page={page} type={type} />
        </Suspense>
      </Container>
    </Box>
  )
}
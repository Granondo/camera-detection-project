import { Box, Container, Heading, Text, SimpleGrid, HStack, Button } from '@chakra-ui/react'
import Link from 'next/link'
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

  const { events, total, total_pages } = await getEvents({
    page,
    perPage: 20,
    type,
  })

  // Build query string for pagination links
  const buildQueryString = (newPage: number) => {
    const query = new URLSearchParams()
    query.set('page', newPage.toString())
    if (type) query.set('type', type)
    return query.toString()
  }

  return (
    <Box minH="100vh">
      <Header />

      <Container maxW="container.xl" py={8}>
        <Box mb={8}>
          <Heading size="xl" mb={2} color="gray.800">
            Detection Events
          </Heading>
          <Text color="gray.600">
            {total} event{total !== 1 ? 's' : ''} found • Page {page} of {total_pages}
          </Text>
        </Box>

        {events.length > 0 ? (
          <>
            <SimpleGrid columns={{ base: 1, sm: 2, lg: 3 }} gap={6} mb={8}>
              {events.map((event) => (
                <EventCard key={event.id} event={event} />
              ))}
            </SimpleGrid>

            {/* Pagination Controls */}
            {total_pages > 1 && (
              <HStack justify="center" gap={2} flexWrap="wrap">
                {/* Previous Button */}
                {page > 1 ? (
                  <Link href={`/events?${buildQueryString(page - 1)}`}>
                    <Button variant="outline" colorPalette="blue" size="sm">
                      Previous
                    </Button>
                  </Link>
                ) : (
                  <Button variant="outline" colorPalette="blue" size="sm" disabled>
                    Previous
                  </Button>
                )}

                {/* Page Numbers */}
                {Array.from({ length: Math.min(total_pages, 5) }, (_, i) => {
                  let pageNum: number
                  if (total_pages <= 5) {
                    pageNum = i + 1
                  } else if (page <= 3) {
                    pageNum = i + 1
                  } else if (page >= total_pages - 2) {
                    pageNum = total_pages - 4 + i
                  } else {
                    pageNum = page - 2 + i
                  }

                  return (
                    <Link key={pageNum} href={`/events?${buildQueryString(pageNum)}`}>
                      <Button
                        variant={page === pageNum ? 'solid' : 'outline'}
                        colorPalette="blue"
                        size="sm"
                      >
                        {pageNum}
                      </Button>
                    </Link>
                  )
                })}

                {/* Next Button */}
                {page < total_pages ? (
                  <Link href={`/events?${buildQueryString(page + 1)}`}>
                    <Button variant="outline" colorPalette="blue" size="sm">
                      Next
                    </Button>
                  </Link>
                ) : (
                  <Button variant="outline" colorPalette="blue" size="sm" disabled>
                    Next
                  </Button>
                )}
              </HStack>
            )}
          </>
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
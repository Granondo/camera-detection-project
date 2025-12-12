import { SimpleGrid, Box, Text, HStack, Button } from '@chakra-ui/react'
import Link from 'next/link'
import { EventCard } from '@/components/EventCard'
import { getEvents } from '@/lib/api'

interface EventsGridProps {
  page: number
  type?: string
}

export async function EventsGrid({ page, type }: EventsGridProps) {
  const { events, total_pages } = await getEvents({
    page,
    perPage: 12,
    type,
  })

  // Build query string for pagination links
  const buildQueryString = (newPage: number) => {
    const query = new URLSearchParams()
    query.set('page', newPage.toString())
    if (type) query.set('type', type)
    return query.toString()
  }

  if (events.length === 0) {
    return (
      <Box textAlign="center" py={12}>
        <Text fontSize="lg" color="gray.500">
          No events found for the selected criteria.
        </Text>
      </Box>
    )
  }

  return (
    <>
      <SimpleGrid columns={{ base: 1, sm: 2, lg: 3, xl: 4 }} gap={6} mb={8}>
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
              <Button variant="outline" size="sm">
                Previous
              </Button>
            </Link>
          ) : (
            <Button variant="outline" size="sm" disabled>
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

            if (pageNum > total_pages) return null

            return (
              <Link key={pageNum} href={`/events?${buildQueryString(pageNum)}`}>
                <Button
                  variant={page === pageNum ? 'solid' : 'outline'}
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
              <Button variant="outline" size="sm">
                Next
              </Button>
            </Link>
          ) : (
            <Button variant="outline" size="sm" disabled>
              Next
            </Button>
          )}
        </HStack>
      )}
    </>
  )
}

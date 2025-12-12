import { SimpleGrid, Box, Text, HStack, Button } from '@chakra-ui/react'
import Link from 'next/link'
import { FrameCard } from '@/components/FrameCard'
import { getFrames } from '@/lib/api'

interface FramesGridProps {
  page: number
  cameraId?: number
}

export async function FramesGrid({ page, cameraId }: FramesGridProps) {
  const { frames, total_pages } = await getFrames({
    page,
    perPage: 16,
    cameraId,
  })

  // Build query string for pagination links
  const buildQueryString = (newPage: number) => {
    const query = new URLSearchParams()
    query.set('page', newPage.toString())
    if (cameraId) query.set('camera_id', cameraId.toString())
    return query.toString()
  }

  if (frames.length === 0) {
    return (
      <Box textAlign="center" py={12}>
        <Text fontSize="lg" color="gray.500">
          No frames found for the selected criteria.
        </Text>
      </Box>
    )
  }

  return (
    <>
      <SimpleGrid columns={{ base: 1, sm: 2, lg: 3, xl: 4 }} gap={6} mb={8}>
        {frames.map((frame) => (
          <FrameCard key={frame.id} frame={frame} />
        ))}
      </SimpleGrid>

      {/* Pagination Controls */}
      {total_pages > 1 && (
        <HStack justify="center" gap={2} flexWrap="wrap">
          {/* Previous Button */}
          {page > 1 ? (
            <Link href={`/frames?${buildQueryString(page - 1)}`}>
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
              <Link key={pageNum} href={`/frames?${buildQueryString(pageNum)}`}>
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
            <Link href={`/frames?${buildQueryString(page + 1)}`}>
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

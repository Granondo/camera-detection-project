import { Box, Container, Heading, Text, SimpleGrid } from '@chakra-ui/react'
import { Header } from '@/components/Header'
import { FrameCard } from '@/components/FrameCard'
import { getFrames } from '@/lib/api'

interface FramesPageProps {
  searchParams: Promise<{
    page?: string
    cameraId?: string
    minConfidence?: string
  }>
}

export const metadata = {
  title: 'Frames | Surveillance',
  description: 'Gallery of frames with detected objects from video surveillance',
}

export default async function FramesPage({ searchParams }: FramesPageProps) {
  const params = await searchParams
  const page = Number(params.page) || 1
  const cameraId = params.cameraId ? Number(params.cameraId) : undefined
  const minConfidence = params.minConfidence ? Number(params.minConfidence) : undefined

  const { frames, total } = await getFrames({
    page,
    perPage: 16,
    cameraId,
    minConfidence,
  })

  return (
    <Box minH="100vh">
      <Header />

      <Container maxW="container.xl" py={8}>
        <Box mb={8}>
          <Heading size="xl" mb={2} color="gray.800">
            Frames Gallery
          </Heading>
          <Text color="gray.600">
            {total} frame{total !== 1 ? 's' : ''} with detections
          </Text>
        </Box>

        {frames.length > 0 ? (
          <SimpleGrid columns={{ base: 1, sm: 2, lg: 3, xl: 4 }} gap={6}>
            {frames.map((frame) => (
              <FrameCard key={frame.id} frame={frame} />
            ))}
          </SimpleGrid>
        ) : (
          <Box textAlign="center" py={12}>
            <Text fontSize="lg" color="gray.500">
              No frames found
            </Text>
          </Box>
        )}
      </Container>
    </Box>
  )
}
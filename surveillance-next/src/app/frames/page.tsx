import { Suspense } from 'react'
import { Box, Container, Heading, Text } from '@chakra-ui/react'
import { Header } from '@/components/Header'
import { GridSkeleton } from '@/components/skeletons/GridSkeleton'
import { FrameCardSkeleton } from '@/components/skeletons/FrameCardSkeleton'
import { FramesGrid } from './FramesGrid'
import { FramesToolbar } from './FramesToolbar'

interface FramesPageProps {
  searchParams: {
    page?: string
    cameraId?: string
  }
}

export const metadata = {
  title: 'Frames | Surveillance',
  description: 'Gallery of frames with detected objects from video surveillance',
}

export default function FramesPage({ searchParams }: FramesPageProps) {
  const page = Number(searchParams?.page) || 1
  const cameraId = searchParams?.cameraId ? Number(searchParams.cameraId) : undefined

  return (
    <Box minH="100vh">
      <Header />

      <Container maxW="container.xl" py={8}>
        <Box mb={4}>
          <Heading size="xl" mb={2} color="gray.800">
            Frames Gallery
          </Heading>
          <Text color="gray.600">
            Browse frames with detected objects from video surveillance.
          </Text>
        </Box>

        <FramesToolbar />

        <Suspense
          key={`${page}-${cameraId || ''}`}
          fallback={
            <GridSkeleton
              count={16}
              skeletonCard={FrameCardSkeleton}
              columns={{ base: 1, sm: 2, lg: 3, xl: 4 }}
            />
          }
        >
          <FramesGrid page={page} cameraId={cameraId} />
        </Suspense>
      </Container>
    </Box>
  )
}
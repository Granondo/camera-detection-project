import { Box, Container, Heading, Text, Badge, SimpleGrid, Flex, Image } from '@chakra-ui/react'
import { format } from 'date-fns'
import { Header } from '@/components/Header'
import { getFrame, getImageUrl } from '@/lib/api'
import type { Metadata } from 'next'

interface FrameDetailPageProps {
  params: Promise<{
    id: string
  }>
}

export async function generateMetadata({ params }: FrameDetailPageProps): Promise<Metadata> {
  const { id } = await params
  const frame = await getFrame(id)
  
  if (!frame) {
    return {
      title: 'Frame Not Found | Surveillance',
    }
  }
  
  return {
    title: `Frame from Camera ${frame.camera_id} | Surveillance`,
    description: `Frame captured at ${frame.timestamp} with ${frame.detections?.length || 0} detections`,
  }
}

export default async function FrameDetailPage({ params }: FrameDetailPageProps) {
  const { id } = await params
  const frame = await getFrame(id)
  
  if (!frame) {
    return (
      <Box minH="100vh">
        <Header />
        <Container maxW="container.xl" py={8}>
          <Text>Frame not found</Text>
        </Container>
      </Box>
    )
  }

  const formattedTime = format(new Date(frame.timestamp), 'dd MMM yyyy, HH:mm:ss')
  const detections = frame.detections || []

  return (
    <Box minH="100vh">
      <Header />
      
      <Container maxW="container.xl" py={8}>
        <Box mb={6}>
          <Heading size="xl" mb={2} color="gray.800">
            Frame Details
          </Heading>
          <Flex gap={4} color="gray.600">
            <Text>📷 Camera {frame.camera_id}</Text>
            <Text>🕐 {formattedTime}</Text>
          </Flex>
        </Box>

        <SimpleGrid columns={{ base: 1, lg: 2 }} gap={8}>
          <Box>
            <Image
              src={getImageUrl(frame.image_path)}
              alt={`Frame ${frame.id}`}
              w="100%"
              borderRadius="lg"
              shadow="md"
            />
          </Box>
          
          <Box>
            <Heading size="md" mb={4} color="gray.800">
              Detections ({detections.length})
            </Heading>
            
            {detections.length > 0 ? (
              <Flex direction="column" gap={3}>
                {detections.map((detection, index) => (
                  <Box
                    key={detection.id || index}
                    bg="white"
                    p={4}
                    borderRadius="md"
                    shadow="sm"
                  >
                    <Flex justify="space-between" align="center" mb={2}>
                      <Badge colorPalette="blue" variant="solid">
                        {detection.type}
                      </Badge>
                      <Badge colorPalette="green" variant="subtle">
                        {(detection.confidence * 100).toFixed(1)}% confidence
                      </Badge>
                    </Flex>
                    
                    <SimpleGrid columns={2} gap={2} fontSize="sm" color="gray.600">
                      <Text>X: {detection.bbox.x.toFixed(0)}</Text>
                      <Text>Y: {detection.bbox.y.toFixed(0)}</Text>
                      <Text>Width: {detection.bbox.width.toFixed(0)}</Text>
                      <Text>Height: {detection.bbox.height.toFixed(0)}</Text>
                    </SimpleGrid>
                  </Box>
                ))}
              </Flex>
            ) : (
              <Box textAlign="center" py={8} bg="white" borderRadius="lg">
                <Text color="gray.500">No detections in this frame</Text>
              </Box>
            )}
          </Box>
        </SimpleGrid>
      </Container>
    </Box>
  )
}
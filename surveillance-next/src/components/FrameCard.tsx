import { Box, Badge, Flex, Text, Image } from '@chakra-ui/react'
import { format } from 'date-fns'
import { getImageUrl, type Frame } from '@/lib/api'

interface FrameCardProps {
  frame: Frame
}

export function FrameCard({ frame }: FrameCardProps) {
  const formattedTime = format(new Date(frame.timestamp), 'HH:mm:ss')
  const detectionsCount = frame.detections.length
  const mainDetection = frame.detections[0]

  return (
    <Box
      bg="white"
      borderRadius="lg"
      shadow="sm"
      overflow="hidden"
      transition="all 0.2s"
      _hover={{ shadow: 'md' }}
    >
      <Box position="relative">
        <Image
          src={getImageUrl(frame.imagePath)}
          alt={`Frame ${frame.id}`}
          w="100%"
          h="180px"
          objectFit="cover"
        />
        {detectionsCount > 0 && (
          <Badge
            position="absolute"
            top={2}
            right={2}
            colorPalette="purple"
            variant="solid"
          >
            {detectionsCount} detection{detectionsCount !== 1 ? 's' : ''}
          </Badge>
        )}
      </Box>

      <Box p={4}>
        <Flex justify="space-between" align="center" mb={2}>
          <Text fontWeight="medium" fontSize="sm" color="gray.800">
            📷 {frame.cameraId}
          </Text>
          <Text fontSize="xs" color="gray.500">
            {formattedTime}
          </Text>
        </Flex>

        {mainDetection && (
          <Flex gap={2} flexWrap="wrap">
            <Badge colorPalette="blue" variant="subtle">
              {mainDetection.type}
            </Badge>
            <Badge colorPalette="green" variant="subtle">
              {(mainDetection.confidence * 100).toFixed(0)}% conf
            </Badge>
          </Flex>
        )}
      </Box>
    </Box>
  )
}
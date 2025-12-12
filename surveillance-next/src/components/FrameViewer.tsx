import { Box, Image, Text, Skeleton } from '@chakra-ui/react'
import { getFrame, getImageUrl, Detection } from '@/lib/api'

interface FrameViewerProps {
  frameId: number
}

function BoundingBox({ detection }: { detection: Detection }) {
  const { bbox, type, confidence } = detection;

  const style = {
    left: `${bbox.x * 100}%`,
    top: `${bbox.y * 100}%`,
    width: `${bbox.width * 100}%`,
    height: `${bbox.height * 100}%`,
  }

  return (
    <Box
      position="absolute"
      style={style}
      border="2px solid"
      borderColor="red.400"
      boxShadow="0 0 5px rgba(255, 0, 0, 0.5)"
    >
      <Text
        as="span"
        bg="red.400"
        color="white"
        fontSize="xs"
        p="2px 4px"
        position="absolute"
        top="-22px"
        left="-2px"
      >
        {type} ({(confidence * 100).toFixed(1)}%)
      </Text>
    </Box>
  )
}

export async function FrameViewer({ frameId }: FrameViewerProps) {
  const frame = await getFrame(frameId.toString())

  if (!frame) {
    return (
      <Box p={4} bg="gray.100" borderRadius="md" textAlign="center">
        <Text color="gray.600">Frame #{frameId} not found.</Text>
      </Box>
    )
  }

  const imageUrl = getImageUrl(frame.image_path)

  return (
    <Box position="relative" w="full" bg="gray.100" borderRadius="md" overflow="hidden">
      <Image
        src={imageUrl}
        alt={`Frame ${frame.id}`}
        w="full"
        h="auto"
        objectFit="contain"
      />
      {frame.detections?.map((detection, index) => (
        <BoundingBox key={index} detection={detection} />
      ))}
    </Box>
  )
}

export function FrameViewerSkeleton() {
  return <Skeleton height="400px" borderRadius="md" />
}
